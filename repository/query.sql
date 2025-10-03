-- name: CreateDailyAppointment :one
insert into appts.daily_appointments (first_name, last_name, appointment_date)
values (sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(appointment_date))
returning *;