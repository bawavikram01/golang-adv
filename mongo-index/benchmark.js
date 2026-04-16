const { MongoClient } = require("mongodb");

const uri = "mongodb://localhost:27017";
const dbName = "testdb";
const collectionName = "users";

const TOTAL_DOCS = 1_000_000; // 1 million

async function run() {
  const client = new MongoClient(uri);
  await client.connect();

  const db = client.db(dbName);
  const collection = db.collection(collectionName);

  console.log("Cleaning old data...");
  await collection.deleteMany({});

  console.log("Inserting documents...");
  const batchSize = 10000;
  let batch = [];

  for (let i = 0; i < TOTAL_DOCS; i++) {
    batch.push({
      name: "user_" + i,
      age: Math.floor(Math.random() * 100),
      email: `user${i}@test.com`,
      createdAt: new Date()
    });

    if (batch.length === batchSize) {
      await collection.insertMany(batch);
      batch = [];
      console.log(`Inserted: ${i}`);
    }
  }

  if (batch.length > 0) {
    await collection.insertMany(batch);
  }

  console.log("✅ Data insertion complete");

  // 🔍 Test WITHOUT index
  console.log("\nRunning query WITHOUT index...");
  let start = Date.now();

  await collection.find({ age: 50 }).toArray();

  let end = Date.now();
  console.log(`Time without index: ${end - start} ms`);

  // 📊 Explain
  let explain1 = await collection.find({ age: 50 }).explain("executionStats");
  console.log("Stage:", explain1.executionStats.executionStages.stage);

  // 🚀 Create index
  console.log("\nCreating index on age...");
  await collection.createIndex({ age: 1 });

  // 🔍 Test WITH index
  console.log("\nRunning query WITH index...");
  start = Date.now();

  await collection.find({ age: 50 }).toArray();

  end = Date.now();
  console.log(`Time with index: ${end - start} ms`);

  // 📊 Explain
  let explain2 = await collection.find({ age: 50 }).explain("executionStats");
  console.log("Stage:", explain2.executionStats.executionStages.stage);

  await client.close();
}

run().catch(console.error);