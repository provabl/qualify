// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

// Store provides persistent storage for the agent using BoltDB
type Store struct {
	db *bbolt.DB
}

// Bucket names
var (
	ConfigBucket      = []byte("config")
	CredentialsBucket = []byte("credentials")
	CacheBucket       = []byte("cache")
)

// New creates a new agent store
func New(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{ConfigBucket, CredentialsBucket, CacheBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the database
func (s *Store) Close() error {
	return s.db.Close()
}

// SetConfig stores a configuration value
func (s *Store) SetConfig(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(ConfigBucket)
		return b.Put([]byte(key), data)
	})
}

// GetConfig retrieves a configuration value
func (s *Store) GetConfig(key string, dest interface{}) error {
	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(ConfigBucket)
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return json.Unmarshal(data, dest)
	})
}

// SetCredential stores AWS credentials (encrypted in production)
// Note: This is a simple implementation. Production should encrypt credentials.
func (s *Store) SetCredential(profile string, creds Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CredentialsBucket)
		return b.Put([]byte(profile), data)
	})
}

// GetCredential retrieves AWS credentials for a profile
func (s *Store) GetCredential(profile string) (*Credentials, error) {
	var creds Credentials
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CredentialsBucket)
		data := b.Get([]byte(profile))
		if data == nil {
			return fmt.Errorf("profile not found: %s", profile)
		}
		return json.Unmarshal(data, &creds)
	})
	if err != nil {
		return nil, err
	}
	return &creds, nil
}

// ListCredentials retrieves all stored credential profiles
func (s *Store) ListCredentials() (map[string]*Credentials, error) {
	profiles := make(map[string]*Credentials)

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CredentialsBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var creds Credentials
			if err := json.Unmarshal(v, &creds); err != nil {
				return fmt.Errorf("unmarshal credentials for %s: %w", k, err)
			}
			profiles[string(k)] = &creds
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// DeleteCredential removes AWS credentials for a profile
func (s *Store) DeleteCredential(profile string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CredentialsBucket)
		return b.Delete([]byte(profile))
	})
}

// SetCache stores a cache entry with optional TTL
func (s *Store) SetCache(key string, value interface{}, ttl time.Duration) error {
	entry := CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CacheBucket)
		return b.Put([]byte(key), data)
	})
}

// GetCache retrieves a cache entry if not expired
func (s *Store) GetCache(key string, dest interface{}) error {
	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CacheBucket)
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key not found: %s", key)
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return fmt.Errorf("unmarshal cache entry: %w", err)
		}

		// Check expiration
		if time.Now().After(entry.ExpiresAt) {
			return fmt.Errorf("cache entry expired: %s", key)
		}

		// Extract the value
		valueData, err := json.Marshal(entry.Value)
		if err != nil {
			return fmt.Errorf("marshal value: %w", err)
		}

		return json.Unmarshal(valueData, dest)
	})
}

// Credentials represents AWS credentials
type Credentials struct {
	AccessKeyID     string    `json:"access_key_id"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token,omitempty"`
	Expiration      time.Time `json:"expiration,omitempty"`
	Region          string    `json:"region,omitempty"`
}

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
}
