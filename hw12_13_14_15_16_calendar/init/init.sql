CREATE USER calendar_user WITH ENCRYPTED PASSWORD 'calendar';
CREATE DATABASE calendar OWNER calendar_user;

\c calendar

CREATE SCHEMA calendar AUTHORIZATION calendar_user;
GRANT ALL ON SCHEMA calendar TO calendar_user;
ALTER ROLE calendar_user SET search_path = calendar, public;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA calendar TO calendar_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA calendar TO calendar_user;