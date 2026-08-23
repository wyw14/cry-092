CREATE TABLE users (
  id text PRIMARY KEY,
  display_name text NOT NULL,
  login text NOT NULL UNIQUE CHECK (login = lower(login)),
  password_hash bytea NOT NULL,
  roles text[] NOT NULL CHECK (cardinality(roles) > 0),
  unit_id text NOT NULL DEFAULT '',
  active boolean NOT NULL DEFAULT true,
  version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL
);
CREATE TABLE refresh_tokens (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id),
  digest bytea NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  version bigint NOT NULL DEFAULT 1 CHECK (version > 0)
);
CREATE INDEX refresh_tokens_user_active_idx ON refresh_tokens(user_id, expires_at) WHERE revoked_at IS NULL;
CREATE TABLE proposals (
  id text PRIMARY KEY,
  representative_id text NOT NULL REFERENCES users(id),
  title text NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
  cause text NOT NULL,
  body text NOT NULL,
  category text NOT NULL,
  attachments jsonb NOT NULL DEFAULT '[]',
  status text NOT NULL CHECK (status IN ('draft','submitted','assigned','accepted','handling','answered','evaluated','archived')),
  submitted_at timestamptz,
  version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX proposals_list_idx ON proposals(status, submitted_at DESC, id);
CREATE INDEX proposals_owner_idx ON proposals(representative_id, created_at DESC);
CREATE TABLE cosponsor_invitations (
  id text PRIMARY KEY,
  proposal_id text NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
  representative_id text NOT NULL REFERENCES users(id),
  status text NOT NULL CHECK (status IN ('pending','accepted','withdrawn')),
  invited_at timestamptz NOT NULL,
  responded_at timestamptz,
  version bigint NOT NULL DEFAULT 1,
  UNIQUE(proposal_id, representative_id)
);
CREATE TABLE proposal_snapshots (
  id text PRIMARY KEY,
  proposal_id text NOT NULL UNIQUE REFERENCES proposals(id),
  title text NOT NULL,
  cause text NOT NULL,
  body text NOT NULL,
  category text NOT NULL,
  cosponsors jsonb NOT NULL,
  submitted_by text NOT NULL REFERENCES users(id),
  submitted_at timestamptz NOT NULL
);
CREATE TABLE assignment_rules (id text PRIMARY KEY, category text NOT NULL, unit_id text NOT NULL, priority integer NOT NULL, active boolean NOT NULL, UNIQUE(category, priority));
CREATE TABLE assignments (
  id text PRIMARY KEY,
  proposal_id text NOT NULL REFERENCES proposals(id),
  unit_id text NOT NULL,
  rule_id text REFERENCES assignment_rules(id),
  status text NOT NULL CHECK (status IN ('pending','assigned','accepted','returned')),
  assigned_by text NOT NULL REFERENCES users(id),
  accepted_by text,
  assigned_at timestamptz NOT NULL,
  accepted_at timestamptz,
  reason text NOT NULL DEFAULT '',
  version bigint NOT NULL DEFAULT 1,
  UNIQUE(proposal_id, id)
);
CREATE INDEX assignments_unit_status_idx ON assignments(unit_id, status, assigned_at DESC);
CREATE TABLE handling_plans (id text PRIMARY KEY, proposal_id text NOT NULL UNIQUE REFERENCES proposals(id), assignment_id text NOT NULL REFERENCES assignments(id), unit_id text NOT NULL, steps text[] NOT NULL, materials jsonb NOT NULL DEFAULT '[]', due_at timestamptz NOT NULL, status text NOT NULL CHECK(status IN ('active','extended','completed')), version bigint NOT NULL DEFAULT 1);
CREATE INDEX handling_plans_due_idx ON handling_plans(status, due_at);
CREATE TABLE extension_requests (id text PRIMARY KEY, plan_id text NOT NULL REFERENCES handling_plans(id), requested_days integer NOT NULL CHECK(requested_days BETWEEN 1 AND 30), reason text NOT NULL, status text NOT NULL CHECK(status IN ('pending','approved','rejected')), requested_by text NOT NULL, reviewed_by text NOT NULL DEFAULT '', requested_at timestamptz NOT NULL, reviewed_at timestamptz, version bigint NOT NULL DEFAULT 1);
CREATE TABLE replies (id text PRIMARY KEY, proposal_id text NOT NULL REFERENCES proposals(id), round integer NOT NULL CHECK(round > 0), kind text NOT NULL CHECK(kind IN ('resolved','planned','explained','reference_only')), summary text NOT NULL, file_ids text[] NOT NULL CHECK(cardinality(file_ids)>0), submitted_by text NOT NULL, submitted_at timestamptz NOT NULL, viewed_at timestamptz, supplements text[] NOT NULL DEFAULT '{}', version bigint NOT NULL DEFAULT 1, UNIQUE(proposal_id, round));
CREATE TABLE evaluations (id text PRIMARY KEY, proposal_id text NOT NULL REFERENCES proposals(id), reply_id text NOT NULL REFERENCES replies(id), round integer NOT NULL, representative_id text NOT NULL REFERENCES users(id), rating text NOT NULL CHECK(rating IN ('satisfied','mostly_satisfied','unsatisfied')), comment text NOT NULL DEFAULT '', created_at timestamptz NOT NULL, UNIQUE(proposal_id, round));
CREATE TABLE rework_rounds (id text PRIMARY KEY, proposal_id text NOT NULL REFERENCES proposals(id), previous_reply_id text NOT NULL REFERENCES replies(id), round integer NOT NULL, unit_id text NOT NULL, reason text NOT NULL, created_at timestamptz NOT NULL, completed_at timestamptz, version bigint NOT NULL DEFAULT 1, UNIQUE(proposal_id, round));
CREATE TABLE supervision_cases (id text PRIMARY KEY, proposal_id text NOT NULL REFERENCES proposals(id), supervisor_id text NOT NULL, level integer NOT NULL CHECK(level BETWEEN 1 AND 3), opinions jsonb NOT NULL DEFAULT '[]', opened_at timestamptz NOT NULL, closed_at timestamptz, version bigint NOT NULL DEFAULT 1);
CREATE TABLE reminders (id text PRIMARY KEY, proposal_id text NOT NULL REFERENCES proposals(id), unit_id text NOT NULL, period_start timestamptz NOT NULL, level text NOT NULL CHECK(level IN ('weekly','overdue','escalated')), attempt integer NOT NULL DEFAULT 0, delivered_at timestamptz, dead_lettered_at timestamptz, last_error text NOT NULL DEFAULT '', created_at timestamptz NOT NULL, UNIQUE(proposal_id, period_start, level));
CREATE TABLE audit_events (id text PRIMARY KEY, aggregate text NOT NULL, aggregate_id text NOT NULL, actor_id text NOT NULL, source text NOT NULL CHECK(source IN ('http','worker','seed')), before_data jsonb, after_data jsonb, reason text NOT NULL, occurred_at timestamptz NOT NULL);
CREATE INDEX audit_events_aggregate_idx ON audit_events(aggregate, aggregate_id, occurred_at DESC);
CREATE TABLE outbox_messages (id text PRIMARY KEY, topic text NOT NULL, payload jsonb NOT NULL, attempts integer NOT NULL DEFAULT 0, available_at timestamptz NOT NULL, delivered_at timestamptz, dead_lettered_at timestamptz, last_error text NOT NULL DEFAULT '', created_at timestamptz NOT NULL);
CREATE INDEX outbox_available_idx ON outbox_messages(available_at, created_at) WHERE delivered_at IS NULL AND dead_lettered_at IS NULL;
CREATE TABLE unit_rankings (unit_id text NOT NULL, period_start timestamptz NOT NULL, period_end timestamptz NOT NULL, period_label text NOT NULL, numerator integer NOT NULL, denominator integer NOT NULL CHECK(denominator >= 0), score_basis_points integer NOT NULL, position integer NOT NULL, PRIMARY KEY(unit_id,period_start,period_end));
CREATE TABLE archive_records (id text PRIMARY KEY, proposal_id text NOT NULL UNIQUE REFERENCES proposals(id), year integer NOT NULL, snapshot_id text NOT NULL REFERENCES proposal_snapshots(id), reply_ids text[] NOT NULL, evaluation_ids text[] NOT NULL, audit_event_ids text[] NOT NULL, archived_at timestamptz NOT NULL, retention_until timestamptz NOT NULL);
