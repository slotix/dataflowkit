CREATE KEYSPACE IF NOT EXISTS dfk
  WITH replication = {'class': 'NetworkTopologyStrategy', 'dfk-east': 1 };

CREATE TABLE IF NOT EXISTS dfk.Intermediate (
  payloadHash text,
  pageID int,
  blockID int,
  fields map<text, text>,
  PRIMARY KEY (payloadHash, pageID, blockID)
) WITH comment = 'Table with intermediary results.';

CREATE TABLE IF NOT EXISTS dfk.Intermediatemaps (
  payloadHash text PRIMARY KEY,
  map text,
) WITH comment = 'Map of intermediary results.';

CREATE TABLE IF NOT EXISTS dfk.Cache (
  key text PRIMARY KEY ,
  value text,
) WITH comment = 'Table with Caches';

CREATE TABLE IF NOT EXISTS dfk.Cookies (
  key text PRIMARY KEY,
  value text,
) WITH comment = 'Table with Cookies';