// MongoDB Demo Data for CashLenX - FOR TESTING ONLY
// This script inserts sample transactions for development and testing
//
// To use this data:
// Option 1 (Manual import to database):
//   docker exec -i cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 --authenticationDatabase admin cashlenx < docker/mongodb/init-mongo-demo.js
//
// Option 2 (Via CLI - Recommended):
//   1. Create Excel file from this data
//   2. Run: cashlenx manage import -i demo-data.xlsx
//
// This keeps demo data separate from production initialization

// Switch to cashlenx database
db = db.getSiblingDB('cashlenx');

print('Loading demo data for CashLenX...');

// Helper function to get date string (YYYY-MM-DD format)
function getDateString(daysAgo) {
  const date = new Date();
  date.setDate(date.getDate() - daysAgo);
  return date.toISOString().split('T')[0];
}

// Get category IDs from existing categories
const categories = {};
db.categories.find({}).forEach(cat => {
  categories[cat.name] = cat._id;
});

print(`Found ${Object.keys(categories).length} categories`);

// Insert demo cash flows
// Aligned with Go CashFlowEntity model:
// - _id: MongoDB ObjectId
// - category_id: reference to category ObjectId
// - belongs_date: date string (YYYY-MM-DD format)
// - flow_type: 'income' or 'expense'
// - amount: numeric amount
// - description: transaction description
// - remark: additional notes
// - create_time: creation timestamp
// - modify_time: last modification timestamp
const cashFlows = [
  // Today's transactions
  {
    _id: ObjectId(),
    category_id: categories['Food & Dining'],
    belongs_date: getDateString(0),
    flow_type: 'expense',
    amount: 45.50,
    description: 'Lunch at Italian restaurant',
    remark: 'Great pasta',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Transportation'],
    belongs_date: getDateString(0),
    flow_type: 'expense',
    amount: 12.00,
    description: 'Uber to office',
    remark: 'Morning commute',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Salary'],
    belongs_date: getDateString(0),
    flow_type: 'income',
    amount: 3500.00,
    description: 'Monthly salary',
    remark: 'Regular income',
    create_time: new Date(),
    modify_time: new Date()
  },
  
  // Yesterday
  {
    _id: ObjectId(),
    category_id: categories['Shopping'],
    belongs_date: getDateString(1),
    flow_type: 'expense',
    amount: 89.99,
    description: 'New shoes',
    remark: 'Sports shoes',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Entertainment'],
    belongs_date: getDateString(1),
    flow_type: 'expense',
    amount: 25.00,
    description: 'Movie tickets',
    remark: 'Watched latest film',
    create_time: new Date(),
    modify_time: new Date()
  },
  
  // This week
  {
    _id: ObjectId(),
    category_id: categories['Utilities'],
    belongs_date: getDateString(3),
    flow_type: 'expense',
    amount: 150.00,
    description: 'Electricity bill',
    remark: 'Monthly bill',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Healthcare'],
    belongs_date: getDateString(4),
    flow_type: 'expense',
    amount: 65.00,
    description: 'Pharmacy',
    remark: 'Prescription refill',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Investment'],
    belongs_date: getDateString(5),
    flow_type: 'income',
    amount: 200.00,
    description: 'Dividend payment',
    remark: 'Stock dividend',
    create_time: new Date(),
    modify_time: new Date()
  },
  
  // Earlier this month
  {
    _id: ObjectId(),
    category_id: categories['Food & Dining'],
    belongs_date: getDateString(10),
    flow_type: 'expense',
    amount: 120.00,
    description: 'Grocery shopping',
    remark: 'Weekly groceries',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Transportation'],
    belongs_date: getDateString(12),
    flow_type: 'expense',
    amount: 45.00,
    description: 'Gas station',
    remark: 'Fill up tank',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Food & Dining'],
    belongs_date: getDateString(15),
    flow_type: 'expense',
    amount: 75.50,
    description: 'Dinner with friends',
    remark: 'Birthday celebration',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Entertainment'],
    belongs_date: getDateString(18),
    flow_type: 'expense',
    amount: 30.00,
    description: 'Concert tickets',
    remark: 'Live music event',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Shopping'],
    belongs_date: getDateString(20),
    flow_type: 'expense',
    amount: 250.00,
    description: 'Clothing',
    remark: 'New wardrobe',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Freelance'],
    belongs_date: getDateString(22),
    flow_type: 'income',
    amount: 500.00,
    description: 'Freelance project',
    remark: 'Web design project',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    category_id: categories['Utilities'],
    belongs_date: getDateString(25),
    flow_type: 'expense',
    amount: 80.00,
    description: 'Internet bill',
    remark: 'Monthly service',
    create_time: new Date(),
    modify_time: new Date()
  }
];

db.cash_flows.insertMany(cashFlows);
print(`Inserted ${cashFlows.length} demo cash flow records`);

// Print summary
const totalIncome = db.cash_flows.aggregate([
  { $match: { flow_type: 'income' } },
  { $group: { _id: null, total: { $sum: '$amount' } } }
]).toArray()[0]?.total || 0;

const totalExpense = db.cash_flows.aggregate([
  { $match: { flow_type: 'expense' } },
  { $group: { _id: null, total: { $sum: '$amount' } } }
]).toArray()[0]?.total || 0;

print('\n=== Demo Data Loaded ===');
print(`Total Income: $${totalIncome.toFixed(2)}`);
print(`Total Expense: $${totalExpense.toFixed(2)}`);
print(`Balance: $${(totalIncome - totalExpense).toFixed(2)}`);
print('========================\n');

print('Demo data loaded successfully!');
