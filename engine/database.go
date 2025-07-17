package engine

func (e *Engine) DBAutoMigrate(v bool) error {
	if !v {
		return nil
	}
	return e.db.AutoMigrate()
}

func (e *Engine) DBDriver() string {
	return e.db.Driver()
}

func (e *Engine) DBVersion() (string, error) {
	return e.db.Version()
}
