-- Migration 000016: tsvector generated columns for full-text search.
-- Adds search_vector columns with triggers for ranking + exact-phrase FTS.
-- ILIKE + trgm indexes (000011) remain for similarity/autocomplete.

-- ─── clients ─────────────────────────────────────────────────────────────────

ALTER TABLE clients ADD COLUMN IF NOT EXISTS search_vector tsvector;

CREATE OR REPLACE FUNCTION clients_search_vector_update() RETURNS trigger AS $$
BEGIN
  new.search_vector :=
    setweight(to_tsvector('simple', coalesce(new.tax_code, '')),         'A') ||
    setweight(to_tsvector('simple', coalesce(new.business_name, '')),    'A') ||
    setweight(to_tsvector('simple', coalesce(new.english_name, '')),     'B') ||
    setweight(to_tsvector('simple', coalesce(new.representative_name, '')), 'C') ||
    setweight(to_tsvector('simple', coalesce(new.industry, '')),         'D');
  RETURN new;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_clients_search_vector ON clients;
CREATE TRIGGER trig_clients_search_vector
  BEFORE INSERT OR UPDATE ON clients
  FOR EACH ROW EXECUTE FUNCTION clients_search_vector_update();

CREATE INDEX IF NOT EXISTS idx_clients_search_vector ON clients USING GIN (search_vector);

UPDATE clients SET search_vector =
  setweight(to_tsvector('simple', coalesce(tax_code, '')),         'A') ||
  setweight(to_tsvector('simple', coalesce(business_name, '')),    'A') ||
  setweight(to_tsvector('simple', coalesce(english_name, '')),     'B') ||
  setweight(to_tsvector('simple', coalesce(representative_name, '')), 'C') ||
  setweight(to_tsvector('simple', coalesce(industry, '')),         'D');

-- ─── engagements ─────────────────────────────────────────────────────────────

ALTER TABLE engagements ADD COLUMN IF NOT EXISTS search_vector tsvector;

CREATE OR REPLACE FUNCTION engagements_search_vector_update() RETURNS trigger AS $$
BEGIN
  new.search_vector :=
    setweight(to_tsvector('simple', coalesce(new.service_type, '')),  'A') ||
    setweight(to_tsvector('simple', coalesce(new.description, '')),   'B');
  RETURN new;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_engagements_search_vector ON engagements;
CREATE TRIGGER trig_engagements_search_vector
  BEFORE INSERT OR UPDATE ON engagements
  FOR EACH ROW EXECUTE FUNCTION engagements_search_vector_update();

CREATE INDEX IF NOT EXISTS idx_engagements_search_vector ON engagements USING GIN (search_vector);

UPDATE engagements SET search_vector =
  setweight(to_tsvector('simple', coalesce(service_type, '')),  'A') ||
  setweight(to_tsvector('simple', coalesce(description, '')),   'B');

-- ─── employees ───────────────────────────────────────────────────────────────

ALTER TABLE employees ADD COLUMN IF NOT EXISTS search_vector tsvector;

CREATE OR REPLACE FUNCTION employees_search_vector_update() RETURNS trigger AS $$
BEGIN
  new.search_vector :=
    setweight(to_tsvector('simple', coalesce(new.full_name, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(new.email, '')),     'B') ||
    setweight(to_tsvector('simple', coalesce(new.position, '')),  'C');
  RETURN new;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_employees_search_vector ON employees;
CREATE TRIGGER trig_employees_search_vector
  BEFORE INSERT OR UPDATE ON employees
  FOR EACH ROW EXECUTE FUNCTION employees_search_vector_update();

CREATE INDEX IF NOT EXISTS idx_employees_search_vector ON employees USING GIN (search_vector);

UPDATE employees SET search_vector =
  setweight(to_tsvector('simple', coalesce(full_name, '')), 'A') ||
  setweight(to_tsvector('simple', coalesce(email, '')),     'B') ||
  setweight(to_tsvector('simple', coalesce(position, '')),  'C');
