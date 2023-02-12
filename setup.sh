#!/usr/local/bin/bash

set -eu

readonly DBFILE_NAME="mygraphql.db"

# Create DB file
if [ ! -e ${DBFILE_NAME} ];then
  echo ".open ${DBFILE_NAME}" | sqlite3
fi

# Create DB Tables
echo "creating tables..."
sqlite3 ${DBFILE_NAME} "
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users(\
	id TEXT PRIMARY KEY NOT NULL,\
	name TEXT NOT NULL,\
	project_v2 TEXT\
);

CREATE TABLE IF NOT EXISTS repositories(\
	id TEXT PRIMARY KEY NOT NULL,\
	owner TEXT NOT NULL,\
	name TEXT NOT NULL,\
	created_at DATETIME NOT NULL DEFAULT (DATETIME('now','localtime')),\
	FOREIGN KEY (owner) REFERENCES users(id)\
);

CREATE TABLE IF NOT EXISTS issues(\
	id TEXT PRIMARY KEY NOT NULL,\
	url TEXT NOT NULL,\
	title TEXT NOT NULL,\
	closed INTEGER NOT NULL DEFAULT 0,\
	number INTEGER NOT NULL,\
	author TEXT NOT NULL,\
	repository TEXT NOT NULL,\
	CHECK (closed IN (0, 1)),\
	FOREIGN KEY (repository) REFERENCES repositories(id),\
	FOREIGN KEY (author) REFERENCES users(id)\
);

CREATE TABLE IF NOT EXISTS projects(\
	id TEXT PRIMARY KEY NOT NULL,\
	title TEXT NOT NULL,\
	url TEXT NOT NULL,\
	number INTEGER NOT NULL,\
	owner TEXT NOT NULL,\
	FOREIGN KEY (owner) REFERENCES users(id)\
);

CREATE TABLE IF NOT EXISTS pullrequests(\
	id TEXT PRIMARY KEY NOT NULL,\
	base_ref_name TEXT NOT NULL,\
	closed INTEGER NOT NULL DEFAULT 0,\
	head_ref_name TEXT NOT NULL,\
	url TEXT NOT NULL,\
	number INTEGER NOT NULL,\
	repository TEXT NOT NULL,\
	CHECK (closed IN (0, 1)),\
	FOREIGN KEY (repository) REFERENCES repositories(id)\
);

CREATE TABLE IF NOT EXISTS projectcards(\
	id TEXT PRIMARY KEY NOT NULL,\
	project TEXT NOT NULL,\
	issue TEXT,\
	pullrequest TEXT,\
	FOREIGN KEY (project) REFERENCES projects(id),\
	FOREIGN KEY (issue) REFERENCES issues(id),\
	FOREIGN KEY (pullrequest) REFERENCES pullrequests(id),\
	CHECK (issue IS NOT NULL OR pullrequest IS NOT NULL)\
);
"

# Insert initial data
echo "inserting initial data..."
sqlite3 ${DBFILE_NAME} "
PRAGMA foreign_keys = ON;

INSERT INTO users(id, name) VALUES\
	('U_1', 'hsaki')
;

INSERT INTO repositories(id, owner, name) VALUES\
	('REPO_1', 'U_1', 'repo1')
;

INSERT INTO issues(id, url, title, closed, number, author, repository) VALUES\
	('ISSUE_1', 'http://example.com/repo1/issue/1', 'First Issue', 1, 1, 'U_1', 'REPO_1'),\
	('ISSUE_2', 'http://example.com/repo1/issue/2', 'Second Issue', 0, 2, 'U_1', 'REPO_1'),\
	('ISSUE_3', 'http://example.com/repo1/issue/3', 'Third Issue', 0, 3, 'U_1', 'REPO_1'),\
	('ISSUE_4', 'http://example.com/repo1/issue/4', '', 0, 4, 'U_1', 'REPO_1'),\
	('ISSUE_5', 'http://example.com/repo1/issue/5', '', 0, 5, 'U_1', 'REPO_1'),\
	('ISSUE_6', 'http://example.com/repo1/issue/6', '', 0, 6, 'U_1', 'REPO_1'),\
	('ISSUE_7', 'http://example.com/repo1/issue/7', '', 0, 7, 'U_1', 'REPO_1')\
;

INSERT INTO projects(id, title, url, number, owner) VALUES\
	('PJ_1', 'My Project', 'http://example.com/project/1', 1, 'U_1'),\
	('PJ_2', 'My Project 2', 'http://example.com/project/2', 2, 'U_1')\
;

INSERT INTO pullrequests(id, base_ref_name, closed, head_ref_name, url, number, repository) VALUES\
	('PR_1', 'main', 1, 'feature/kinou1', 'http://example.com/repo1/pr/1', 1, 'REPO_1'),\
	('PR_2', 'main', 0, 'feature/kinou2', 'http://example.com/repo1/pr/2', 2, 'REPO_1')\
;
"
