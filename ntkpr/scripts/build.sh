cd ..
if [ -f ./bin/ntkpr ]; then
    rm ./bin/ntkpr
fi
go build -o ./bin/ntkpr
./bin/ntkpr