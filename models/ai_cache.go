package models

// AICache stores AI-classified subcategory results so they survive restarts.
type AICache struct {
	ID            int64   `db:"id,pk autoincr"`
	InputText     string  `db:"input_text,uqs"`
	SubcategoryID string  `db:"subcategory_id"`
	Intent        string  `db:"intent"`
	Confidence    float64 `db:"confidence"`
	CreatedAt     int64   `db:"created_at"`
}

func (AICache) TableName() string {
	return "ai_cache"
}
