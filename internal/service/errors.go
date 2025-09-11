package service

import "errors"

// Common service layer errors - centralized for consistency and reusability
var (
	// Generic resource errors
	ErrNotFound = errors.New("resource not found")

	// Thread-related errors
	ErrThreadNotFound       = errors.New("thread not found")
	ErrThreadAlreadyExists  = errors.New("thread already exists")
	ErrInvalidThreadID      = errors.New("invalid thread ID")
	ErrThreadStatusInvalid  = errors.New("invalid thread status")
	ErrOptimisticLockFailed = errors.New("optimistic lock failed - resource was modified")

	// Mention-related errors
	ErrMentionNotFound      = errors.New("mention not found")
	ErrMentionAlreadyExists = errors.New("mention already exists for this user and thread")
	ErrInvalidMentionID     = errors.New("invalid mention ID")

	// Bot Cookie-related errors
	ErrBotCookieNotFound = errors.New("bot cookie not found")
	ErrInvalidCookieData = errors.New("invalid cookie data")
	ErrCookieExpired     = errors.New("cookie has expired")

	// Processed Mark-related errors
	ErrProcessedMarkNotFound      = errors.New("processed mark not found")
	ErrProcessedMarkAlreadyExists = errors.New("processed mark already exists")

	// Scraping-related errors
	ErrScrapingFailed    = errors.New("scraping failed")
	ErrScrapingTimeout   = errors.New("scraping timeout")
	ErrInvalidTwitterURL = errors.New("invalid Twitter URL")
	ErrTweetNotFound     = errors.New("tweet not found")
	ErrThreadEmpty       = errors.New("thread contains no tweets")

	// IPFS-related errors
	ErrIPFSStoreFailed = errors.New("failed to store content in IPFS")
	ErrIPFSLoadFailed  = errors.New("failed to load content from IPFS")
	ErrInvalidCID      = errors.New("invalid IPFS CID")

	// LLM-related errors
	ErrLLMGenerationFailed = errors.New("LLM generation failed")
	ErrLLMTimeout          = errors.New("LLM generation timeout")
	ErrEmptyLLMResponse    = errors.New("empty LLM response")

	// Permission-related errors
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrOnlyOwnerCanEdit   = errors.New("only the owner can edit this resource")
	ErrOnlyOwnerCanDelete = errors.New("only the owner can delete this resource")

	// Validation errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrMissingParameter = errors.New("missing required parameter")
	ErrInvalidFormat    = errors.New("invalid format")
)
