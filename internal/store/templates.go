package store

import "context"

// GetTemplate повертає (template, isCustom). isCustom=true означає, що
// користувач задав свій шаблон (а не використовується дефолт із events.DefaultTemplates).
func (s *Store) GetTemplate(ctx context.Context, event, kind string) (string, bool) {
	var t string
	if err := s.pool.QueryRow(ctx,
		`SELECT template FROM monitor_templates WHERE event = $1 AND kind = $2`, event, kind).Scan(&t); err != nil {
		return "", false
	}
	return t, true
}

func (s *Store) SetTemplate(ctx context.Context, event, kind, tpl string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO monitor_templates (event, kind, template) VALUES ($1, $2, $3)
		ON CONFLICT (event, kind) DO UPDATE SET template = EXCLUDED.template, updated_at = now()`,
		event, kind, tpl)
	return err
}

func (s *Store) DeleteTemplate(ctx context.Context, event, kind string) (bool, error) {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM monitor_templates WHERE event = $1 AND kind = $2`, event, kind)
	return tag.RowsAffected() > 0, err
}

func (s *Store) ListTemplates(ctx context.Context) ([]Template, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT event, kind, template FROM monitor_templates ORDER BY event, kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Template
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.Event, &t.Kind, &t.Template); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
