-- Insert sample banners for testing
INSERT INTO banners (id, name, created_at, updated_at, is_active) VALUES
(1, 'Banner 1 - Main Page', NOW(), NOW(), true),
(2, 'Banner 2 - Sidebar', NOW(), NOW(), true),
(3, 'Banner 3 - Footer', NOW(), NOW(), true),
(4, 'Banner 4 - Header', NOW(), NOW(), true),
(5, 'Banner 5 - Mobile', NOW(), NOW(), true),
(6, 'Banner 6 - Desktop', NOW(), NOW(), true),
(7, 'Banner 7 - Tablet', NOW(), NOW(), true),
(8, 'Banner 8 - Popup', NOW(), NOW(), true),
(9, 'Banner 9 - Newsletter', NOW(), NOW(), true),
(10, 'Banner 10 - Special Offer', NOW(), NOW(), true)
ON CONFLICT (id) DO NOTHING;

-- Reset sequence to continue from the highest ID
SELECT setval('banners_id_seq', (SELECT MAX(id) FROM banners)); 