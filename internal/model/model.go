package model

// AllModels returns a slice of all models that need to be migrated
// Order is important: tables with foreign keys should come after the tables they reference
func AllModels() []any {
	return []any{
		// First create base tables without foreign keys
		&UserProfile{},
		&ProcessedMention{},
		&BotCookie{},
		// Add new models here as they are created
	}
}
