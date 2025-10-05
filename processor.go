package main

import (
	"sync/atomic"
	"time"

	"log/slog"

	"github.com/cooldogedev/spectrum/session"
	"github.com/oomph-ac/oomph/player"
	"github.com/oomph-ac/oomph/player/component"
	"github.com/oomph-ac/oomph/player/context"
	"github.com/oomph-ac/oomph/player/detection"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type OomphProcessor struct {
	session.NopProcessor

	identity login.IdentityData
	registry *session.Registry
	log      *slog.Logger
	pl       atomic.Pointer[player.Player]

	done chan struct{}
}

func NewOomphProcessor(
	s *session.Session,
	registry *session.Registry,
	listener *minecraft.Listener,
	log *slog.Logger,
) *OomphProcessor {
	pl := player.New(log, player.MonitoringState{
		IsReplay:    false,
		IsRecording: false,
		CurrentTime: time.Now(),
	}, listener)
	pl.SetConn(s.Client())
	component.Register(pl)
	detection.Register(pl)
	oomphProcessor := &OomphProcessor{
		identity: s.Client().IdentityData(),
		registry: registry,

		log:  log,
		done: make(chan struct{}),
	}
	oomphProcessor.pl.Store(pl)
	go pl.StartTicking()

	playerIdentifier := pl.Conn().IdentityData().XUID
	if playerIdentifier == "" {
		playerIdentifier = pl.Conn().ClientData().SelfSignedID
	}
	return oomphProcessor
}

func (p *OomphProcessor) ProcessStartGame(ctx *session.Context, gd *minecraft.GameData) {
	//gd.PlayerMovementSettings.MovementType = protocol.PlayerMovementModeServerWithRewind
	gd.PlayerMovementSettings.RewindHistorySize = 40
}

func (p *OomphProcessor) ProcessServer(ctx *session.Context, pk *packet.Packet) {
	pl := p.pl.Load()
	if pl == nil {
		return
	}

	pkCtx := context.NewHandlePacketContext(pk)
	pl.HandleServerPacket(pkCtx)
	if pkCtx.Cancelled() {
		ctx.Cancel()
	}
}

func (p *OomphProcessor) ProcessClient(ctx *session.Context, pk *packet.Packet) {
	pl := p.pl.Load()
	if pl == nil || pl.Conn() == nil {
		return
	}

	pkCtx := context.NewHandlePacketContext(pk)
	pl.HandleClientPacket(pkCtx)
	if pkCtx.Cancelled() {
		ctx.Cancel()
	}
}

func (p *OomphProcessor) ProcessFlush(ctx *session.Context) {
	if pl := p.pl.Load(); pl != nil {
		pl.PauseProcessing()
		defer pl.ResumeProcessing()

		conn := pl.Conn()
		if conn == nil {
			return
		}

		// We want to handle flushing manually to ensure that we can flush the ACKs while the player lock is held.
		ctx.Cancel()

		pl.ACKs().Flush()
		if err := conn.Flush(); err != nil {
			pl.Log().Error("error flushing client connection", "error", err)
		}
	}
}

func (p *OomphProcessor) ProcessPreTransfer(*session.Context, *string, *string) {
	if pl := p.pl.Load(); pl != nil {
		pl.PauseProcessing()
	}
}

func (p *OomphProcessor) ProcessPostTransfer(_ *session.Context, _ *string, _ *string) {
	if s, pl := p.registry.GetSession(p.identity.XUID), p.pl.Load(); s != nil && pl != nil {
		pl.SetServerConn(s.Server())
		pl.ResumeProcessing()
	}
}

func (p *OomphProcessor) ProcessDisconnection(_ *session.Context, _ *string) {
	if pl := p.pl.Load(); pl != nil {
		_ = pl.Close()
		for id := range pl.EntityTracker().All() {
			pl.EntityTracker().RemoveEntity(id)
		}
		for id := range pl.ClientEntityTracker().All() {
			pl.ClientEntityTracker().RemoveEntity(id)
		}
		pl.World().SetSTWTicks(300)
		pl.Effects().RemoveAll()
		p.pl.Store(nil)
		close(p.done)
	}
}

func (p *OomphProcessor) Player() *player.Player {
	return p.pl.Load()
}
