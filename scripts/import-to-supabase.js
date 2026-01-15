const { createClient } = require('@supabase/supabase-js');
const fs = require('fs');

const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

if (!SUPABASE_URL || !SUPABASE_SERVICE_KEY) {
    console.error('❌ Error: Missing environment variables');
    console.error('   Please set SUPABASE_URL and SUPABASE_SERVICE_KEY');
    console.error('\nUsage:');
    console.error('  SUPABASE_URL=https://xxx.supabase.co SUPABASE_SERVICE_KEY=xxx node import-to-supabase.js');
    process.exit(1);
}

const supabase = createClient(SUPABASE_URL, SUPABASE_SERVICE_KEY);

async function importData() {
    console.log('📦 Loading export data...\n');

    if (!fs.existsSync('linky-export.json')) {
        console.error('❌ Error: linky-export.json not found');
        console.error('   Please run the export script first:');
        console.error('   go run export-sqlite-to-json.go');
        process.exit(1);
    }

    const data = JSON.parse(fs.readFileSync('linky-export.json', 'utf8'));

    // Ensure all arrays exist (even if empty)
    data.tags = data.tags || [];
    data.link_tags = data.link_tags || [];

    const userIdMap = {};
    const subjectIdMap = {};
    const topicIdMap = {};
    const linkIdMap = {};
    const tagIdMap = {};

    console.log('📊 Data to import:');
    console.log(`   - ${data.users.length} users`);
    console.log(`   - ${data.subjects.length} subjects`);
    console.log(`   - ${data.topics.length} topics`);
    console.log(`   - ${data.links.length} links`);
    console.log(`   - ${data.tags.length} tags`);
    console.log(`   - ${data.link_tags.length} link_tags\n`);

    // Step 1: Create users in Supabase Auth
    console.log('👤 Step 1/6: Creating users in Supabase Auth...');
    for (const user of data.users) {
        const tempPassword = 'TempPass_' + Math.random().toString(36).substring(2, 15);
        const { data: authUser, error } = await supabase.auth.admin.createUser({
            email: user.email,
            password: tempPassword,
            email_confirm: true
        });

        if (error) {
            // User already exists, try to fetch them
            if (error.message.includes('already been registered')) {
                console.log(`   ℹ️  User ${user.email} already exists, fetching...`);

                // List users and find by email
                const { data: users, error: listError } = await supabase.auth.admin.listUsers();
                if (!listError && users) {
                    const existingUser = users.users.find(u => u.email === user.email);
                    if (existingUser) {
                        userIdMap[user.id] = existingUser.id;
                        console.log(`   ✓ ${user.email} -> ${existingUser.id} (existing)`);
                        continue;
                    }
                }
            }
            console.error(`   ❌ Error with user ${user.email}:`, error.message);
            continue;
        }

        userIdMap[user.id] = authUser.user.id;
        console.log(`   ✓ ${user.email} -> ${authUser.user.id} (new)`);
    }
    console.log(`   ✅ Processed ${Object.keys(userIdMap).length} users\n`);

    // Step 2: Import subjects
    console.log('📁 Step 2/6: Importing subjects...');
    for (const subject of data.subjects) {
        if (!userIdMap[subject.user_id]) {
            console.log(`   ⚠️  Skipping subject "${subject.name}" - user not found`);
            continue;
        }

        const { data: newSubject, error } = await supabase
            .from('subjects')
            .insert({
                user_id: userIdMap[subject.user_id],
                name: subject.name
            })
            .select()
            .single();

        if (error) {
            console.error(`   ❌ Error inserting subject "${subject.name}":`, error.message);
            continue;
        }

        subjectIdMap[subject.id] = newSubject.id;
    }
    console.log(`   ✅ Imported ${Object.keys(subjectIdMap).length} subjects\n`);

    // Step 3: Import topics
    console.log('📋 Step 3/6: Importing topics...');
    for (const topic of data.topics) {
        if (!subjectIdMap[topic.subject_id]) {
            console.log(`   ⚠️  Skipping topic "${topic.name}" - subject not found`);
            continue;
        }

        const { data: newTopic, error } = await supabase
            .from('topics')
            .insert({
                subject_id: subjectIdMap[topic.subject_id],
                name: topic.name
            })
            .select()
            .single();

        if (error) {
            console.error(`   ❌ Error inserting topic "${topic.name}":`, error.message);
            continue;
        }

        topicIdMap[topic.id] = newTopic.id;
    }
    console.log(`   ✅ Imported ${Object.keys(topicIdMap).length} topics\n`);

    // Step 4: Import links (in batches of 100)
    console.log('🔗 Step 4/6: Importing links...');
    let importedLinks = 0;

    for (let i = 0; i < data.links.length; i += 100) {
        const batch = data.links.slice(i, i + 100);
        const linksToInsert = batch
            .filter(link => topicIdMap[link.topic_id])
            .map(link => ({
                topic_id: topicIdMap[link.topic_id],
                title: link.title,
                url: link.url
            }));

        if (linksToInsert.length === 0) continue;

        const { data: newLinks, error } = await supabase
            .from('links')
            .insert(linksToInsert)
            .select();

        if (error) {
            console.error(`   ❌ Error inserting links batch:`, error.message);
            continue;
        }

        if (newLinks) {
            const validBatch = batch.filter(link => topicIdMap[link.topic_id]);
            validBatch.forEach((oldLink, idx) => {
                linkIdMap[oldLink.id] = newLinks[idx].id;
            });
            importedLinks += newLinks.length;
        }

        process.stdout.write(`   Progress: ${importedLinks}/${data.links.length} links...\r`);
    }
    console.log(`\n   ✅ Imported ${Object.keys(linkIdMap).length} links\n`);

    // Step 5: Import tags
    console.log('🏷️  Step 5/6: Importing tags...');
    for (const tag of data.tags) {
        if (!userIdMap[tag.user_id]) {
            console.log(`   ⚠️  Skipping tag "${tag.name}" - user not found`);
            continue;
        }

        const { data: newTag, error } = await supabase
            .from('tags')
            .insert({
                user_id: userIdMap[tag.user_id],
                name: tag.name
            })
            .select()
            .single();

        if (error) {
            console.error(`   ❌ Error inserting tag "${tag.name}":`, error.message);
            continue;
        }

        tagIdMap[tag.id] = newTag.id;
    }
    console.log(`   ✅ Imported ${Object.keys(tagIdMap).length} tags\n`);

    // Step 6: Import link_tags
    console.log('🔖 Step 6/6: Importing link_tags...');
    const linkTagsToInsert = data.link_tags
        .filter(lt => linkIdMap[lt.link_id] && tagIdMap[lt.tag_id])
        .map(lt => ({
            link_id: linkIdMap[lt.link_id],
            tag_id: tagIdMap[lt.tag_id]
        }));

    if (linkTagsToInsert.length > 0) {
        const { error } = await supabase
            .from('link_tags')
            .insert(linkTagsToInsert);

        if (error) {
            console.error('   ❌ Error inserting link_tags:', error.message);
        } else {
            console.log(`   ✅ Imported ${linkTagsToInsert.length} link_tags\n`);
        }
    } else {
        console.log('   ℹ️  No link_tags to import\n');
    }

    console.log('═══════════════════════════════════════════════');
    console.log('✅ Migration completed successfully!');
    console.log('═══════════════════════════════════════════════\n');
    console.log('📋 Summary:');
    console.log(`   Users:     ${Object.keys(userIdMap).length}/${data.users.length}`);
    console.log(`   Subjects:  ${Object.keys(subjectIdMap).length}/${data.subjects.length}`);
    console.log(`   Topics:    ${Object.keys(topicIdMap).length}/${data.topics.length}`);
    console.log(`   Links:     ${Object.keys(linkIdMap).length}/${data.links.length}`);
    console.log(`   Tags:      ${Object.keys(tagIdMap).length}/${data.tags.length}`);
    console.log(`   Link_tags: ${linkTagsToInsert.length}/${data.link_tags.length}`);
    console.log('\n⚠️  IMPORTANT: Users will need to reset their passwords!');
    console.log('   Temporary passwords were used during migration.');
    console.log('   Users should use "Forgot Password" to set new passwords.\n');
}

importData().catch(error => {
    console.error('\n❌ Fatal error during migration:', error);
    process.exit(1);
});
