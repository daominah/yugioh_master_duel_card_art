package yugioh_master_duel_card_art

import (
	"testing"
)

func TestReadAllCardDataKonami(t *testing.T) {
	cards := ReadAllCardDataKonami()
	if len(cards) < 1000 {
		t.Fatalf("error unexpected number of cards: %v", len(cards))
	}

	if got, want := cards["4007"].CardName, "Blue-Eyes White Dragon"; got != want {
		t.Fatalf("error unexpected card English name: %v, want: %v", got, want)
	}
	if got, want := NormalizeName(cards["4007"].CardName), "blue_eyes_white_dragon"; got != want {
		t.Fatalf("error NormalizeName: %v, want: %v", got, want)
	}

	if got, want := cards["3423"].CardName, "Eldlich the Golden Lord"; got != want {
		t.Fatalf("error unexpected card English name: %v, want: %v", got, want)
	}
	if got, want := cards["3423"].MonsterATK, 2500.0; got != want {
		t.Fatalf("error NormalizeName: %v, want: %v", got, want)
	}

	// Test alt arts: Blue-Eyes White Dragon (4007) should have alt art 3801
	if cards["3801"].CardID != "4007" {
		t.Fatalf("error alt art 3801 should map to original card 4007")
	}
	if cards["3801"].AltArtID != "3801" {
		t.Fatalf("error alt art 3801 should have AltArtID = 3801")
	}
	if cards["3801"].CardName != "Blue-Eyes White Dragon" {
		t.Fatalf("error alt art 3801 should have same name as original")
	}

	if cards["21230"].CardID != "17785" {
		t.Fatalf("error alt art 21230 should map to original card 17785")
	}
	if cards["21230"].AltArtID != "21230" {
		t.Fatalf("error alt art 21230 should have AltArtID = 21230")
	}
	if cards["21230"].CardName != "Lady Labrynth of the Silver Castle" {
		t.Fatalf("error alt art 21230 should have same name as original")
	}

	// Test cards have 3+ alt arts
	for _, cardID := range []string{"14676", "3415", "22746"} {
		if cards[cardID].CardName != "I:P Masquerena" {
			t.Fatalf("error card %v should be I:P Masquerena", cardID)
		}
	}
}
