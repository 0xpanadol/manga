DELETE FROM roles_permissions WHERE permission_id = (SELECT id FROM permissions WHERE code = 'chapters:manage');
DELETE FROM permissions WHERE code = 'chapters:manage';

DROP TABLE IF EXISTS "chapters";