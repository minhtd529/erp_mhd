package ws_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mdh/erp-audit/api/pkg/ws"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// ── Hub registration ─────────────────────────────────────────────────────────

func TestHub_RegisterAndCount(t *testing.T) {
	t.Parallel()
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	if got := hub.ClientCount(); got != 0 {
		t.Fatalf("want 0 clients, got %d", got)
	}

	c1, peer1 := makeClient(t, hub, []string{"global"})
	defer peer1.Close()
	go c1.ReadPump()
	go c1.WritePump()

	// Give the hub goroutine time to process the register event.
	time.Sleep(20 * time.Millisecond)

	if got := hub.ClientCount(); got != 1 {
		t.Fatalf("want 1 client, got %d", got)
	}
	if got := hub.SubscriberCount("global"); got != 1 {
		t.Fatalf("want 1 subscriber on 'global', got %d", got)
	}
	if got := hub.SubscriberCount("crm"); got != 0 {
		t.Fatalf("want 0 subscribers on 'crm', got %d", got)
	}
}

func TestHub_UnregisterOnClose(t *testing.T) {
	t.Parallel()
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	c1, peer1 := makeClient(t, hub, []string{"global"})
	go c1.ReadPump()
	go c1.WritePump()

	time.Sleep(20 * time.Millisecond)
	if hub.ClientCount() != 1 {
		t.Fatal("client not registered")
	}

	// Close the peer side — ReadPump should detect the error and unregister.
	peer1.Close()
	time.Sleep(50 * time.Millisecond)

	if got := hub.ClientCount(); got != 0 {
		t.Fatalf("want 0 clients after close, got %d", got)
	}
}

// ── Broadcast ────────────────────────────────────────────────────────────────

func TestHub_BroadcastDeliveredToSubscriber(t *testing.T) {
	t.Parallel()
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	c1, peer1 := makeClient(t, hub, []string{"crm"})
	defer peer1.Close()
	go c1.ReadPump()
	go c1.WritePump()

	time.Sleep(20 * time.Millisecond)

	// Send an event on the "crm" channel.
	if err := hub.Broadcast("crm", "crm.client.created", map[string]string{"id": "abc123"}); err != nil {
		t.Fatalf("Broadcast error: %v", err)
	}

	// Read the message from the peer (browser-side connection).
	peer1.SetReadDeadline(time.Now().Add(500 * time.Millisecond)) //nolint:errcheck
	_, raw, err := peer1.ReadMessage()
	if err != nil {
		t.Fatalf("expected message from hub, got error: %v", err)
	}

	var msg ws.Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}
	if msg.Type != "crm.client.created" {
		t.Errorf("want type %q, got %q", "crm.client.created", msg.Type)
	}
	if msg.Channel != "crm" {
		t.Errorf("want channel %q, got %q", "crm", msg.Channel)
	}
}

func TestHub_BroadcastNotDeliveredToOtherChannel(t *testing.T) {
	t.Parallel()
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Client subscribes to "global", not "crm".
	c1, peer1 := makeClient(t, hub, []string{"global"})
	defer peer1.Close()
	go c1.ReadPump()
	go c1.WritePump()

	time.Sleep(20 * time.Millisecond)

	hub.Broadcast("crm", "crm.client.created", nil) //nolint:errcheck

	// Peer should receive nothing (tight timeout).
	peer1.SetReadDeadline(time.Now().Add(80 * time.Millisecond)) //nolint:errcheck
	_, _, err := peer1.ReadMessage()
	if err == nil {
		t.Fatal("expected no message on 'global' client for 'crm' broadcast")
	}
}

func TestHub_MultipleChannelSubscription(t *testing.T) {
	t.Parallel()
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Client subscribes to both channels.
	c1, peer1 := makeClient(t, hub, []string{"global", "crm"})
	defer peer1.Close()
	go c1.ReadPump()
	go c1.WritePump()

	time.Sleep(20 * time.Millisecond)

	if hub.SubscriberCount("global") != 1 {
		t.Error("expected subscriber on 'global'")
	}
	if hub.SubscriberCount("crm") != 1 {
		t.Error("expected subscriber on 'crm'")
	}
}

// ── parseChannels (via handler) ───────────────────────────────────────────────

func TestParseChannels(t *testing.T) {
	// Tested indirectly through ws.Handler; here we test the exported behaviour
	// by verifying channel filtering at the hub level.
	t.Parallel()
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	c1, peer1 := makeClient(t, hub, ws.ParseChannels("global,hrm"))
	defer peer1.Close()
	go c1.ReadPump()
	go c1.WritePump()
	time.Sleep(20 * time.Millisecond)

	if hub.SubscriberCount("global") != 1 {
		t.Error("expected subscriber on global")
	}
	if hub.SubscriberCount("hrm") != 1 {
		t.Error("expected subscriber on hrm")
	}
	if hub.SubscriberCount("crm") != 0 {
		t.Error("expected no subscriber on crm")
	}
}
