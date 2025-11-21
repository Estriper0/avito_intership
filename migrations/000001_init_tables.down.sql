DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS pull_requests_statuses;

DROP INDEX IF EXISTS idx_users_team_id;
DROP INDEX IF EXISTS idx_pr_author_id;
DROP INDEX IF EXISTS idx_pr_reviewers_pr_id;
DROP INDEX IF EXISTS idx_pr_reviewers_user_id;