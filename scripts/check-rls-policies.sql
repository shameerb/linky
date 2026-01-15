-- Check and Fix RLS Policies for Linky App
-- Run this in Supabase SQL Editor

-- Check if RLS is enabled on tables
SELECT
    tablename,
    rowsecurity
FROM
    pg_tables
WHERE
    schemaname = 'public'
    AND tablename IN ('subjects', 'topics', 'links', 'tags', 'link_tags');

-- View existing policies on links table
SELECT
    schemaname,
    tablename,
    policyname,
    permissive,
    roles,
    cmd,
    qual,
    with_check
FROM
    pg_policies
WHERE
    tablename = 'links';

-- If RLS policies are missing, run the following:

-- Enable RLS on all tables (if not already enabled)
ALTER TABLE subjects ENABLE ROW LEVEL SECURITY;
ALTER TABLE topics ENABLE ROW LEVEL SECURITY;
ALTER TABLE links ENABLE ROW LEVEL SECURITY;
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE link_tags ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist (to recreate them)
DROP POLICY IF EXISTS "Users can delete own links" ON links;

-- Recreate the delete policy for links
CREATE POLICY "Users can delete own links" ON links
    FOR DELETE USING (
        EXISTS (
            SELECT 1 FROM topics
            JOIN subjects ON topics.subject_id = subjects.id
            WHERE topics.id = links.topic_id AND subjects.user_id = auth.uid()
        )
    );

-- Verify the policy was created
SELECT
    policyname,
    cmd,
    qual
FROM
    pg_policies
WHERE
    tablename = 'links'
    AND cmd = 'DELETE';
