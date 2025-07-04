DELETE FROM roles_permissions WHERE permission_id = (SELECT id FROM permissions WHERE code = 'manga:manage');
DELETE FROM permissions WHERE code = 'manga:manage';

DROP TABLE IF EXISTS "manga_genres";
DROP TABLE IF EXISTS "manga";
DROP TABLE IF EXISTS "genres";
DROP TYPE IF EXISTS manga_status;

