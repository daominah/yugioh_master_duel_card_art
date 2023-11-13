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
}
