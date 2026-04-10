-- Idempotent seed data for local demo and reviewer flows.
-- Primary credential:
--   email: test@example.com
--   password: password123

INSERT INTO users (id, name, email, password)
VALUES
  (
    '11111111-1111-1111-1111-111111111111',
    'Test User',
    'test@example.com',
    '$2a$12$JUSkqV7vppybjiwHX1yNI.aE2/vSNRDKuDAw2yeL9PJI2cyjNcP8.'
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'Demo Collaborator',
    'collab@example.com',
    '$2a$12$JUSkqV7vppybjiwHX1yNI.aE2/vSNRDKuDAw2yeL9PJI2cyjNcP8.'
  )
ON CONFLICT (email) DO UPDATE
SET
  name = EXCLUDED.name,
  password = EXCLUDED.password;

INSERT INTO projects (id, name, description, owner_id)
VALUES
  (
    '33333333-3333-3333-3333-333333333333',
    'TaskFlow Demo Project',
    'Seeded project for reviewer walkthroughs',
    '11111111-1111-1111-1111-111111111111'
  )
ON CONFLICT (id) DO UPDATE
SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  owner_id = EXCLUDED.owner_id;

INSERT INTO tasks (
  id,
  title,
  description,
  status,
  priority,
  project_id,
  assignee_id,
  created_by,
  due_date
)
VALUES
  (
    '44444444-4444-4444-4444-444444444444',
    'Draft initial API shape',
    'Create initial endpoint payload contracts',
    'todo',
    'high',
    '33333333-3333-3333-3333-333333333333',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    CURRENT_DATE + 7
  ),
  (
    '55555555-5555-5555-5555-555555555555',
    'Set up migration runner',
    'Wire migration execution into startup flow',
    'in_progress',
    'medium',
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    CURRENT_DATE + 4
  ),
  (
    '66666666-6666-6666-6666-666666666666',
    'Prepare reviewer smoke-test notes',
    'List expected demo credentials and checks',
    'done',
    'low',
    '33333333-3333-3333-3333-333333333333',
    NULL,
    '11111111-1111-1111-1111-111111111111',
    NULL
  )
ON CONFLICT (id) DO UPDATE
SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  status = EXCLUDED.status,
  priority = EXCLUDED.priority,
  project_id = EXCLUDED.project_id,
  assignee_id = EXCLUDED.assignee_id,
  created_by = EXCLUDED.created_by,
  due_date = EXCLUDED.due_date,
  updated_at = now();
