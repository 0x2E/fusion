-- Auto mark-read preference (#202): when the reader marks an opened article as
-- read, and after how long. A closed set validated in the handler: "off" (the
-- client default when NULL), "open" (immediately on open), or a delay in
-- seconds ("5", "10", "30"). No DEFAULT: NULL must stay distinguishable from
-- an explicit choice, and a default here would silently change behavior for
-- every existing database on upgrade.

ALTER TABLE settings ADD COLUMN auto_mark_read TEXT;
