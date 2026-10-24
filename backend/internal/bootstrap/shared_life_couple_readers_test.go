package bootstrap

import (
	"context"
	"testing"
)

func TestSharedLifeCoupleReadersTranslateAtCompositionRoot(t *testing.T) {
	service := stubCurrentCouple{}
	anniversary, err := newAnniversaryCoupleReader(service).GetActiveCouple(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	todo, err := newTodoCoupleReader(service).GetActiveCouple(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	wish, err := newWishCoupleReader(service).GetActiveCouple(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if anniversary.ID != 12 || todo.ID != 12 || wish.ID != 12 {
		t.Fatalf("anniversary=%+v todo=%+v wish=%+v", anniversary, todo, wish)
	}
}
