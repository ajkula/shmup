package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	config.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configurer la gestion des signaux pour une fermeture propre
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	g, err := game.NewGame(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(config.Config.ScreenWidth, config.Config.ScreenHeight)
	ebiten.SetWindowTitle("Shmup Game")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
