if ! which git > /dev/null; then
    echo "Git is required to install dependencies. Please install Git and then try again."
    exit 1
fi

if ! which go > /dev/null; then
    echo "Go is not installed. Please install Go and then try again."
    exit 1
fi

if [ ! -d "deps/oomph" ]; then
    git clone https://github.com/oomph-ac/oomph deps/oomph
else
    cd deps/oomph; git pull; cd ../..;
fi

if [ ! -d "deps/dragonfly" ]; then
    git clone https://github.com/oomph-ac/dragonfly deps/dragonfly
else
    cd deps/dragonfly; git pull; cd ../..;
fi

if [ ! -d "deps/oconfig" ]; then
    git clone https://github.com/oomph-ac/oconfig deps/oconfig
else
    cd deps/oconfig; git pull; cd ../..;
fi

if [ ! -d "deps/spectrum" ]; then
    git clone https://github.com/oomph-ac/spectrum deps/spectrum
else
    cd deps/spectrum; git pull; cd ../..;
fi

if [ ! -d "deps/gophertunnel" ]; then
    git clone https://github.com/sandertv/gophertunnel deps/gophertunnel
else
    cd deps/gophertunnel; git pull; cd ../..;
fi

go get
go mod tidy
