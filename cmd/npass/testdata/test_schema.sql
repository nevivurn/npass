-- Initialize schema
CREATE TABLE keys (
	id	INTEGER	PRIMARY KEY NOT NULL,
	name	TEXT	UNIQUE NOT NULL,
	public	TEXT	NOT NULL,
	private	TEXT
);
CREATE TABLE pass (
	id	INTEGER	PRIMARY KEY NOT NULL,
	key_id	INTEGER	NOT NULL REFERENCES keys(id),
	name	TEXT	NOT NULL,
	type	TEXT	NOT NULL,
	data	TEXT	NOT NULL,
	UNIQUE	(key_id, name)
);
CREATE TABLE meta (
	key	TEXT	PRIMARY KEY NOT NULL,
	value	TEXT	NOT NULL
);
INSERT INTO meta (key, value) VALUES('version', '1');

-- Insert test data

-- Test keys
INSERT INTO keys (id, name, public, private) VALUES
-- pass: test-1
(1, 'test-1', 'CM4ICfq6RLZ/P6qqKElNe5Pr+pk+v1PKJrbTzsbvSHk', '3pgVPY99/aKrbxy5823b+oybwiszOvR2xI26DZo/EK2CoRBdlYi9b/RXQJkYNyJvvEiJ3vShSuVfuW7XoAhGtQ'),
-- pass: test-2
(2, 'test-2', 'LzjhStmiT786jQslhaHcREWoy9vwGOvDqfXHVTfZfxY', 'DIfeLde/lWgEYXMtKN1p3kAXTvihBWTBbaAT6t4Kz0nBtRDeWSYlJuA2HjTd82AFRVnOZIymNtdgmEAXjy7zmw');
