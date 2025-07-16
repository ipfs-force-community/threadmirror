package model

// AllModels returns a slice of all models that need to be migrated
// Order is important: tables with foreign keys should come after the tables they reference
func AllModels() []any {
	return []any{
		// First create base tables without foreign keys
		&ProcessedMark{},
		&BotCookie{},
		&Mention{},
		&Thread{},
		&Translation{}, // Translation table references Thread
		// Add new models here as they are created
	}
}
