# SQL initialisation

The sql-files in this folder initialize the Nomen database and user. These objects need to exist before Nomen is able to set and start up.

## files

- 01_user.sql: create the user Nomen uses to connect to the database
- 02_database.sql: create the database for Nomen
- 03_grant_user.sql: grant the runtime user the DDL and DML access Nomen migrations require
- 04_eventstore.sql: creates the schema needed for eventsourcing
- 05_projections.sql: creates the schema needed to read the data
- 06_system.sql: creates the schema needed for Nomen itself
- 07_encryption_keys_table.sql: creates the table for encryption keys (for event data)
- 08_events_table.sql creates the table for eventsourcing
- 09_nomen_schema.sql creates the identity schema; tables are applied by migrations
- 10_unique_constraints_table.sql creates the table to check unique constraints for events
- 11_nomen_product_schema.sql creates the product overlay schema; tables are applied by migrations
