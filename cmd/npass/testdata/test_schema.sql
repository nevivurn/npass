BEGIN;

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
	UNIQUE	(key_id, name, type)
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
(1, 'test-1', '5M60s3mDoFQJQOkZxyn0nOo2VuKzdUJZU60j3Ymkny4', 'iPKo4J3hZNL3yOHjTFez1FWaax96HPlGVG3azFMLBHrfnkV4D4WZMQ2dQaYw6n/BxmxlVSsVGMqgH1niHKP3qg'),
-- pass: test-2
(2, 'test-2', 'VIFZnL4uDjJIwU61yIN0NV5ORYdehWU2PYeTwMwDvwc', 'ALEHXVTZkQx3X0OS3aVoH2E/TvYIjJ6bb3gvlkVvEhA1+qcbVDXI3d/Q/XCuU4VfsQbKbf374KEccP1yp2ymmg');

-- Test passwords
INSERT INTO pass (id, key_id, name, type, data) VALUES
-- pass: pass-1
(1, 1, 'test-1', 'pass', 'LUgYJuptQvsht2iIrKJ9tOnAPyG4V3XYKsCyda59/ly5iYCWejMYBQyN5lt0Nf6G4TelWpAAQTKDhrrfQPD3hxHUbxXsYuVdkg'),
-- pass: pass-2
(2, 2, 'test-2', 'pass', 'ktsKjAoBVw55Y7dU5GKocjPDxVQDeI+yMpEDPu5eBwnYTzSJ7JqlWcMT79EJlbuN/1Ht1P+OqAmiR1phDLEMbViyFDsq1GxNgg');

COMMIT;
