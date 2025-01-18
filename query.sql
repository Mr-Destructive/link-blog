-- name: CreateLink :execlastid
INSERT INTO links (url, commentary) VALUES (?, ?);

-- name: GetLink :one
SELECT id, url, commentary FROM links WHERE id = ?;

-- name: GetLinksByURL :many
SELECT id, url, commentary FROM links WHERE url = ?;

-- name: ListLinks :many
SELECT id, url, commentary FROM links ORDER BY id;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = ?;

-- name: UpdateLink :exec
UPDATE links SET url = ?, commentary = ? WHERE id = ?;
