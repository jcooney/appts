create SCHEMA appts;

create user appt_user with PASSWORD 'appt_password';

grant USAGE on SCHEMA appts TO appt_user;

create TABLE IF NOT EXISTS appts.daily_appointments (
    ID integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_name varchar(50) NOT NULL,
    last_name varchar(50) NOT NULL,
    appointment_date timestamp with time zone NOT NULL
);

grant select, insert, update, delete on ALL TABLES IN SCHEMA appts TO appt_user;

create unique index unique_appointment_day on appts.daily_appointments (appointment_date);