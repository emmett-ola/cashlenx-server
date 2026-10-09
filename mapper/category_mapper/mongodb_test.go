package category_mapper

import (
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConvertBsonM2CategoryEntityKeepsMissingResultEmpty(t *testing.T) {
	converted := convertBsonM2CategoryEntity(bson.M{})

	if !converted.IsEmpty() {
		t.Fatalf("missing result converted to non-empty category: %#v", converted)
	}
}

func TestConvertBsonM2CategoryEntityBackfillsLegacyPresentation(t *testing.T) {
	converted := convertBsonM2CategoryEntity(bson.M{
		"_id":             primitive.NewObjectID(),
		"belongs_user_id": primitive.NewObjectID(),
		"name":            "Legacy",
		"type":            "expense",
	})

	if converted.Emoji != model.DefaultCategoryEmoji {
		t.Fatalf("emoji = %q, want %q", converted.Emoji, model.DefaultCategoryEmoji)
	}
	if converted.BgColor != model.DefaultCategoryBackgroundColor {
		t.Fatalf("bg_color = %q, want %q", converted.BgColor, model.DefaultCategoryBackgroundColor)
	}
}
