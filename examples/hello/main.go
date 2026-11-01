// Command hello is the smallest whole program: it joins the first LAN world it
// finds, says what it sees, and leaves. Run it first, to check that everything
// up to the world works: the LAN search, the sign-in (only if the world asks
// for one) and the joining.
//
//	go run github.com/otya0/mcbedrock-bot/examples/hello@latest
//
// Before that, open a world in Minecraft with "Visible to LAN Players" on, and
// take Mojang's data once (see the README).
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/otya0/mcbedrock-bot"
)

func main() {
	// Give up after two minutes; a first sign-in, if the world asks for one,
	// counts in them too.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// The logger shows each step of joining, so a failure says where it stopped.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bot, err := mcbedrock.Join(ctx, mcbedrock.Options{Logger: logger})
	if err != nil {
		log.Fatalf("join: %v", err)
	}
	defer bot.Close()
	fmt.Println("joined")

	// Wait a moment for the world around the bot to arrive, then look.
	time.Sleep(3 * time.Second)
	health, maximum := bot.Self().Health()
	fmt.Printf("at %.1f, health %.0f of %.0f, %d entities around\n",
		bot.Self().Position(), health, maximum, len(bot.Entities().All()))

	// Stay ten seconds, then leave.
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("leaving")
	case <-bot.Done():
		fmt.Println("the connection ended:", bot.Err())
	}
}
