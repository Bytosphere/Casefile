CREATE TABLE T_Issue (
     id          INTEGER PRIMARY KEY AUTOINCREMENT,
     title       TEXT    NOT NULL,
     description TEXT,
     severity    TEXT    NOT NULL CHECK (severity IN ('Low', 'Medium', 'High', 'Critical')),
     file        TEXT    NOT NULL,
     line        INTEGER NOT NULL,
     status      TEXT    NOT NULL CHECK (status IN ('Open', 'Closed')),
     created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
     fingerprint TEXT    NOT NULL
);

INSERT INTO T_Issue (title, description, severity, file, line, status, created_at, fingerprint)
VALUES
    ('SQL Injection Risk', 'User input is not properly sanitized before being used in SQL query', 'Critical', 'internal/handler/user.go', 127, 'Open', '2024-01-15T10:30:00Z', 'sql_injection_user_handler'),
    ('Missing Error Check', 'Function return value is not checked for errors', 'High', 'internal/service/auth.go', 45, 'Open', '2024-01-16T14:22:00Z', 'unchecked_error_auth'),
    ('Hardcoded Credential', 'Password or API key is hardcoded in source code', 'Critical', 'internal/config/keys.go', 12, 'Open', '2024-01-17T09:15:00Z', 'hardcoded_credential'),
    ('Unclosed Resource', 'File handle is not closed after use', 'Medium', 'internal/io/reader.go', 88, 'Closed', '2024-01-18T11:45:00Z', 'unclosed_file'),
    ('Potential Nil Pointer', 'Variable may be nil before dereference', 'High', 'internal/core/processor.go', 203, 'Open', '2024-01-19T16:30:00Z', 'nil_deref_processor'),
    ('Deprecated Function', 'Using deprecated API that will be removed in future version', 'Low', 'internal/utils/parser.go', 34, 'Open', '2024-01-20T08:00:00Z', 'deprecated_api_parser'),
    ('Missing Validation', 'Input parameter is not validated', 'Medium', 'internal/handler/api.go', 67, 'Closed', '2024-01-21T13:20:00Z', 'missing_input_validation'),
    ('Race Condition', 'Concurrent access to shared resource without synchronization', 'High', 'internal/cache/store.go', 156, 'Open', '2024-01-22T10:10:00Z', 'race_condition_store');
