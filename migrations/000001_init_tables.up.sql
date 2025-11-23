CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(30) PRIMARY KEY,
    username VARCHAR(30) NOT NULL,
    team_id INTEGER NOT NULL REFERENCES teams(id),
    is_active BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS pull_requests_statuses (
    id INTEGER PRIMARY KEY,
    status VARCHAR(30) UNIQUE NOT NULL
);

INSERT INTO pull_requests_statuses (id, status) VALUES
    (1, 'OPEN'),
    (2, 'MERGED');

CREATE TABLE IF NOT EXISTS pull_requests (
    pr_id VARCHAR(30) PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    author_id VARCHAR(30) NOT NULL REFERENCES users(user_id),
    status_id INTEGER NOT NULL DEFAULT 1 REFERENCES pull_requests_statuses(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS pull_requests_reviewers (
    pr_id VARCHAR(30) REFERENCES pull_requests(pr_id),
    user_id VARCHAR(30) REFERENCES users(user_id),
    PRIMARY KEY(pr_id, user_id)
);

CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX idx_pr_reviewers_pr_id ON pull_requests_reviewers(pr_id);
CREATE INDEX idx_pr_reviewers_user_id ON pull_requests_reviewers(user_id);