package database

// GetSetting retrieves a setting value by key.
func (db *DB) GetSetting(key string) (string, error) {
	db.WaitForReady()
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting stores a setting value.
func (db *DB) SetSetting(key, value string) error {
	db.WaitForReady()
	_, err := db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}
