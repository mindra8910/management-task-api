-- =====================================================
-- SAMPLE DATA FOR TASK MANAGEMENT API
-- =====================================================
-- Password untuk semua user sample: "password123"
-- Hash menggunakan bcrypt dengan cost 10
-- =====================================================

-- Sample Users
INSERT INTO users (id, name, email, password, created_at) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'John Doe', 'john@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW()),
    ('550e8400-e29b-41d4-a716-446655440002', 'Jane Smith', 'jane@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW()),
    ('550e8400-e29b-41d4-a716-446655440003', 'Bob Wilson', 'bob@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW())
ON CONFLICT (id) DO NOTHING;

-- Sample Tasks untuk John Doe (user1)
INSERT INTO tasks (id, user_id, title, description, status, created_at, updated_at) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'Setup project', 'Initialize Go project with dependencies', 'completed', NOW(), NOW()),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'Implement API', 'Create REST API endpoints', 'in_progress', NOW(), NOW()),
    ('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'Write tests', 'Unit testing for usecases', 'pending', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Sample Tasks untuk Jane Smith (user2)
INSERT INTO tasks (id, user_id, title, description, status, created_at, updated_at) VALUES
    ('660e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', 'Design database', 'Design PostgreSQL schema', 'completed', NOW(), NOW()),
    ('660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440002', 'Implement auth', 'JWT authentication middleware', 'in_progress', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Sample Task Logs
INSERT INTO task_logs (task_id, action, created_at) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'Task created', NOW()),
    ('660e8400-e29b-41d4-a716-446655440001', 'Status changed to completed', NOW()),
    ('660e8400-e29b-41d4-a716-446655440002', 'Task created', NOW()),
    ('660e8400-e29b-41d4-a716-446655440004', 'Task created', NOW())
ON CONFLICT DO NOTHING;
