CREATE TABLE T_Item (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO T_Item
VALUES
    (1, 'Hello, World!'),
    (2, 'A'),
    (3, 'B'),
    (4, 'C'),
    (5, 'D');

DELETE FROM T_Item WHERE name = 'B';