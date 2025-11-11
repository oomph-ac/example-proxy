package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/cooldogedev/spectrum"
	"github.com/cooldogedev/spectrum/server"
	"github.com/cooldogedev/spectrum/session"
	"github.com/cooldogedev/spectrum/util"
	"github.com/getsentry/sentry-go"
	"github.com/oomph-ac/oconfig"
	"github.com/oomph-ac/oomph/oerror"
	"github.com/oomph-ac/oomph/player"
	"github.com/oomph-ac/oomph/utils"
	"github.com/oomph-ac/oomph/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sandertv/gophertunnel/minecraft/text"

	"net/http"
	_ "net/http/pprof"

	_ "github.com/oomph-ac/oomph"
)

var (
	proxy      *spectrum.Spectrum
	moderators = make(map[string]struct{})
)

func init() {
	// Configure slog with text handler
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           os.Getenv("SENTRY_DSN"),
		EnableTracing: false,
		Debug:         false,
	}); err != nil {
		slog.Error("failed to initialize sentry", "error", err)
		os.Exit(1)
	}

	f, err := os.OpenFile("moderators.list", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("failed to read moderators list", "error", err)
		os.Exit(1)
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		moderators[line] = struct{}{}
	}
}

func main() {
	if err := oconfig.ParseJSON("oomph_config.hjson"); err != nil {
		slog.Error("unable to parse config", "error", err)
		os.Exit(1)
	}

	/* if codeSec[0] == 0x1 {
		os.Exit(1)
	} */

	_ = os.Mkdir("./logs", 0644)
	if os.Getenv("PPROF_ENABLED") != "" {
		go http.ListenAndServe(os.Getenv("PPROF_ADDRESS"), nil)
	}

	gcPct := oconfig.Global.GCPercent
	if gcPct < 100 && gcPct != -1 {
		slog.Warn("GCPercent is set to a value which will cause high CPU usage. We have automatically adjusted this value back to 100. Please refer to https://tip.golang.org/doc/gc-guide for more information.")
		gcPct = 100
	}

	debug.SetGCPercent(gcPct)
	debug.SetMemoryLimit(int64(oconfig.Global.MemThreshold) * 1024 * 1024) // Convert MB to bytes

	opts := util.DefaultOpts()
	opts.ClientDecode = player.ClientDecode
	opts.AutoLogin = false
	opts.Addr = oconfig.Global.LocalAddress
	opts.SyncProtocol = false
	opts.ShutdownMessage = oconfig.Global.ShutdownMessage
	//opts.Token = oconfig.Global.SpectrumKey

	statusProvider, err := minecraft.NewForeignStatusProvider(oconfig.Global.RemoteAddress)
	if err != nil {
		panic(err)
	}

	packs := []*resource.Pack{}
	if oconfig.Global.Resource.ResourceFolder != "" {
		newPacks, err := utils.ResourcePacks(oconfig.Global.Resource.ResourceFolder, "content_keys.json")
		if err != nil {
			slog.Error("unable to load resource packs", "error", err)
			time.Sleep(3 * time.Second)
		} else {
			packs = newPacks
		}
	}

	// Register custom blocks here before calling these two functions.
	world.FinalizeBlockRegistry()
	utils.InitializeBlockNameMapping()

	proxy = spectrum.NewSpectrum(
		server.NewStaticDiscovery(oconfig.Global.RemoteAddress, oconfig.Global.BackupAddress),
		slog.Default(),
		opts,
		nil,
	)

	if err := proxy.Listen(minecraft.ListenConfig{
		ResourcePacks:        packs,
		TexturePacksRequired: oconfig.Global.Resource.RequirePacks,
		StatusProvider:       statusProvider,
		FlushRate:            -1,
		//AcceptedProtocols:    legacyver.All(false),
	}); err != nil {
		slog.Error("unable to listen", "address", oconfig.Global.LocalAddress, "error", err)
		os.Exit(1)
	}

	go handleConnections()
	slog.Info("Oomph proxy is running", "address", oconfig.Global.LocalAddress)

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	<-interruptChan

	address, port, validBackupAddr := "", uint16(0), false
	if oconfig.Global.BackupAddress != "" {
		split := strings.Split(oconfig.Global.BackupAddress, ":")
		if len(split) != 2 {
			slog.Warn("invalid backup address - expected two items when splitting address and port", "address", oconfig.Global.BackupAddress)
		} else {
			address = split[0]
			prt, err := strconv.ParseInt(split[1], 10, 16)
			if err != nil {
				slog.Warn("invalid backup address - expected a valid port number", "address", oconfig.Global.BackupAddress)
			} else {
				port = uint16(prt)
				validBackupAddr = true
			}
		}
	}

	for _, s := range proxy.Registry().GetSessions() {
		if !validBackupAddr {
			s.Disconnect(oconfig.Global.ShutdownMessage)
		} else {
			s.Client().WritePacket(&packet.Transfer{
				Address:     address,
				Port:        port,
				ReloadWorld: false,
			})
			s.Client().Flush()
		}
	}
}

func handleConnections() {
	defer func() {
		if v := recover(); v != nil {
			hub := sentry.CurrentHub().Clone()
			sentryErrID := hub.Recover(oerror.New("handleConnections goroutine crashed: %v", v))
			_ = hub.Flush(time.Second * 10)

			// Because we're in production and don't want the whole proxy program to crash, we will just restart the goroutine.
			slog.Warn("handleConnections goroutine crashed", "errorID", *sentryErrID)
			go handleConnections()
		}
	}()

	for {
		s, err := proxy.Accept()
		if err != nil {
			slog.Error("failed to accept session", "error", err)
			continue
		}
		go acceptSession(s)
	}
}

func acceptSession(s *session.Session) {
	defer func(xuid string) {
		if v := recover(); v != nil {
			hub := sentry.CurrentHub().Clone()
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("xuid", xuid)
			})
			sentryErrID := hub.Recover(oerror.New("acceptSession goroutine crashed: %v", v))
			_ = hub.Flush(time.Second * 10)
			slog.Warn("acceptSession goroutine crashed", "errorID", *sentryErrID)
			s.Disconnect(text.Colourf("<red><bold>An error occured while processing your connection.</bold></red>\nError ID: %s", *sentryErrID))
		}
	}(s.Client().IdentityData().XUID)

	// Disable auto-login so that Oomph's processor can modify the StartGame data to allow server-authoritative movement.
	f, err := os.OpenFile(fmt.Sprintf("./logs/%s.log", s.Client().IdentityData().DisplayName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
	if err != nil {
		s.Disconnect("failed to create log file")
		return
	}

	playerLog := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	proc := NewOomphProcessor(s, proxy.Registry(), proxy.Listener(), playerLog)
	pl := proc.Player()
	pl.SetRecoverFunc(recoverPlayerFn)
	pl.HandleEvents(oomphHandler)
	if _, ok := moderators[pl.Name()]; ok {
		pl.AddPerm(player.PermissionAlerts)
		pl.AddPerm(player.PermissionLogs)
		pl.AddPerm(player.PermissionDebug)
	}
	s.SetProcessor(proc)

	if err := s.Login(); err != nil {
		s.Disconnect(err.Error())
		if !errors.Is(err, context.Canceled) {
			slog.Error("session failed to login", "error", err)
		}
		return
	}
	proc.Player().SetServerConn(s.Server())
}
