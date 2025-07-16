#!/usr/bin/env node

// Auth0 Configuration Checker for ThreadMirror Extension

const fs = require('fs');
const path = require('path');

console.log('🔍 ThreadMirror Extension - Auth0 Configuration Checker\n');

// Check if .env.local exists
const envPath = path.join(__dirname, '.env.local');
if (!fs.existsSync(envPath)) {
  console.log('❌ .env.local file not found');
  console.log('📋 Please run: cp env.example .env.local');
  console.log('📖 Then edit .env.local with your Auth0 credentials\n');
  process.exit(1);
}

// Read .env.local file
const envContent = fs.readFileSync(envPath, 'utf8');
const envVars = {};

envContent.split('\n').forEach(line => {
  line = line.trim();
  if (line && !line.startsWith('#')) {
    const [key, value] = line.split('=');
    if (key && value) {
      envVars[key.trim()] = value.trim();
    }
  }
});

console.log('✅ .env.local file found');

// Check required variables
const checks = [
  {
    key: 'VITE_AUTH0_DOMAIN',
    name: 'Auth0 Domain',
    required: true,
    validate: (value) => {
      if (!value || value === 'your-domain.auth0.com') {
        return 'Not configured - please set your actual Auth0 domain';
      }
      if (!value.includes('.auth0.com') && !value.includes('.')) {
        return 'Invalid format - should be like: your-domain.auth0.com';
      }
      return null;
    }
  },
  {
    key: 'VITE_AUTH0_CLIENT_ID',
    name: 'Auth0 Client ID',
    required: true,
    validate: (value) => {
      if (!value || value === 'your-auth0-client-id') {
        return 'Not configured - please set your actual Auth0 Client ID';
      }
      if (value.length < 20) {
        return 'Too short - Auth0 Client IDs are usually 30+ characters';
      }
      return null;
    }
  },
  {
    key: 'VITE_AUTH0_AUDIENCE',
    name: 'Auth0 Audience',
    required: false,
    validate: () => null
  },
  {
    key: 'VITE_API_BASE_URL',
    name: 'API Base URL',
    required: false,
    validate: () => null
  }
];

let hasErrors = false;

console.log('\n📋 Configuration Check Results:');
console.log('================================');

checks.forEach(check => {
  const value = envVars[check.key];
  const error = check.validate(value);
  
  if (check.required && !value) {
    console.log(`❌ ${check.name}: Missing`);
    hasErrors = true;
  } else if (error) {
    console.log(`❌ ${check.name}: ${error}`);
    hasErrors = true;
  } else if (value) {
    const displayValue = check.key === 'VITE_AUTH0_CLIENT_ID' 
      ? value.substring(0, 8) + '...' 
      : value;
    console.log(`✅ ${check.name}: ${displayValue}`);
  } else {
    console.log(`⚠️  ${check.name}: Not set (optional)`);
  }
});

console.log('\n' + '='.repeat(50));

if (hasErrors) {
  console.log('\n❌ Configuration has errors!');
  console.log('\n🔧 To fix:');
  console.log('1. Edit .env.local file');
  console.log('2. Get credentials from https://manage.auth0.com/');
  console.log('3. Run: npm run build');
  console.log('4. Reload extension in Chrome');
  console.log('\n📖 See AUTH0_SETUP.md for detailed instructions');
  process.exit(1);
} else {
  console.log('\n✅ Configuration looks good!');
  console.log('\n🚀 Next steps:');
  console.log('1. Run: npm run build');
  console.log('2. Load extension in Chrome');
  console.log('3. Test the login functionality');
  console.log('\n💡 If you still get errors, check AUTH0_SETUP.md');
} 