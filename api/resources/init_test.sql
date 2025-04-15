INSERT INTO organizations (id, name, address)
VALUES
    (
        '86e6a1f3-d7aa-4e74-a20a-ea78bc13340b',
        'The Lowell General Hospital',
        '295 Varnum Ave'
    ),
    (
        'a6868293-b590-44f9-bf7e-1381beaf17d6',
        'Worcester Outpatient Clinic',
        '605 Lincoln Street'
    );

INSERT INTO users (id, username, password, role)
VALUES
    (
        '0b6f13da-efb9-4221-9e89-e2729ae90030',
        'jdoe',
        '$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6',
        'patient'
    ),
    (
        'e2c66717-12bb-4b6a-b7b6-3be939e170ad',
        'ghouse',
        '$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6',
        'doctor'
    ),
    (
        '99ab51c4-a544-4352-a8df-4632ff8b105d',
        'admin',
        '$2a$12$FDfWu4JA9ABiG3JmSLTiKOzYn6/5UmXydNpkMssqt/9d47tqhQLX6',
        'admin'
    );
