# BS Open Replay Go parser

[Beat Saber Open Replay format](https://github.com/BeatLeader/BS-Open-Replay) parser written in Go

**Disclaimer**: This is my Go learning project, so expect bugs and ugly code

## Install

```
go get -u github.com/motzel/go-bsor
```

## Usage

```go
path := "replays/hellfire.bsor"

file, err := os.Open(path)
if err != nil {
    log.Fatal("Can not open file: ", err)
}

defer file.Close()

var replay bsor.Bsor

err = bsor.Read(*file, &replay)
if err != nil {
    log.Fatal("Read error: ", err)
}

fmt.Printf("BSOR version: %v\n", replay.Header.Version)
fmt.Printf("BSOR Info: %+v\n", replay.Info)
fmt.Printf("BSOR Frames: %v\n", len(replay.Frames))
fmt.Printf("BSOR Notes: %v\n", len(replay.Notes))
fmt.Printf("BSOR Walls: %v\n", len(replay.Walls))
fmt.Printf("BSOR Heights: %v\n", len(replay.Heights))
fmt.Printf("BSOR Pauses: %v\n", len(replay.Pauses))
```
