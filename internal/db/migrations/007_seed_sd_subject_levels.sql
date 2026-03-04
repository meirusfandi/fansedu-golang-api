-- SD: Matematika, IPA, IPS (sama seperti SMP)
INSERT INTO subject_levels (subject_id, level_id, sort_order)
SELECT s.id, l.id, row_number() OVER (ORDER BY s.sort_order)::int
FROM subjects s
CROSS JOIN levels l
WHERE l.slug = 'sd' AND s.slug IN ('matematika', 'ipa', 'ips')
ON CONFLICT (subject_id, level_id) DO NOTHING;
