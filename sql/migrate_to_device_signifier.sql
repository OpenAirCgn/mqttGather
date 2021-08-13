
	CREATE TABLE IF NOT EXISTS device (
		device_id        INTEGER PRIMARY KEY AUTOINCREMENT,
		device_signifier VARCHAR UNIQUE -- this is the MAC addr of the openoise device
	);

INSERT INTO device (device_signifier) SELECT DISTINCT client FROM dba_stats;

ALTER TABLE dba_stats ADD COLUMN device_id INTEGER REFERENCES device(device_id);
ALTER TABLE tele_mem  ADD COLUMN device_id INTEGER REFERENCES device(device_id);
ALTER TABLE tele_ver  ADD COLUMN device_id INTEGER REFERENCES device(device_id);
ALTER TABLE tele_misc ADD COLUMN device_id INTEGER REFERENCES device(device_id);

UPDATE dba_stats SET device_id = (SELECT device_id FROM device WHERE device_signifier = dba_stats.client);
UPDATE tele_mem  SET device_id = (SELECT device_id FROM device WHERE device_signifier = tele_mem.client);
UPDATE tele_ver  SET device_id = (SELECT device_id FROM device WHERE device_signifier = tele_ver.client);
UPDATE tele_misc SET device_id = (SELECT device_id FROM device WHERE device_signifier = tele_misc.client);

-- you'll need sqlite 3.35 or higher for this.
ALTER TABLE dba_stats DROP COLUMN client;
ALTER TABLE tele_mem DROP COLUMN client;
ALTER TABLE tele_ver DROP COLUMN client;
ALTER TABLE tele_misc DROP COLUMN client;
