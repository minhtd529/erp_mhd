DROP TRIGGER IF EXISTS trig_clients_search_vector ON clients;
DROP FUNCTION IF EXISTS clients_search_vector_update();
DROP INDEX IF EXISTS idx_clients_search_vector;
ALTER TABLE clients DROP COLUMN IF EXISTS search_vector;

DROP TRIGGER IF EXISTS trig_engagements_search_vector ON engagements;
DROP FUNCTION IF EXISTS engagements_search_vector_update();
DROP INDEX IF EXISTS idx_engagements_search_vector;
ALTER TABLE engagements DROP COLUMN IF EXISTS search_vector;

DROP TRIGGER IF EXISTS trig_employees_search_vector ON employees;
DROP FUNCTION IF EXISTS employees_search_vector_update();
DROP INDEX IF EXISTS idx_employees_search_vector;
ALTER TABLE employees DROP COLUMN IF EXISTS search_vector;
