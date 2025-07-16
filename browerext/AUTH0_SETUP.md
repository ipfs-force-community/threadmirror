# Auth0 Setup Guide for ThreadMirror Extension

## 🚀 Quick Setup

### Step 1: Create Auth0 Application

1. Go to [Auth0 Dashboard](https://manage.auth0.com/)
2. Click **Applications** → **Create Application**
3. Choose **Single Page Application**
4. Give it a name like "ThreadMirror Extension"

### Step 2: Configure Application Settings

In your Auth0 application settings:

1. **Allowed Callback URLs**: 
   ```
   https://*.chromiumapp.org/
   ```

2. **Allowed Logout URLs**:
   ```
   https://*.chromiumapp.org/
   ```

3. **Allowed Web Origins**:
   ```
   https://*.chromiumapp.org
   ```

4. **Allowed Origins (CORS)**:
   ```
   https://*.chromiumapp.org
   ```

### Step 3: Get Your Credentials

From your Auth0 application **Settings** tab, copy:
- **Domain** (e.g., `your-domain.auth0.com`)
- **Client ID** (long string like `abc123...xyz789`)

### Step 4: Configure Environment Variables

1. Copy `env.example` to `.env.local`:
   ```bash
   cp env.example .env.local
   ```

2. Edit `.env.local` and replace the placeholders:
   ```bash
   VITE_AUTH0_DOMAIN=your-domain.auth0.com
   VITE_AUTH0_CLIENT_ID=your-actual-client-id
   ```

### Step 5: Build and Test

1. Build the extension:
   ```bash
   npm run build
   ```

2. Load the extension in Chrome and test login

## 🔧 Troubleshooting

### "Authorization page could not be loaded"

This error usually means:

1. **❌ Wrong Auth0 Domain**
   - Check your domain in `.env.local`
   - Should be like: `your-domain.auth0.com`
   - Verify it exists at https://manage.auth0.com/

2. **❌ Wrong Client ID**
   - Copy the exact Client ID from Auth0 dashboard
   - It should be a long string (30+ characters)

3. **❌ Missing Callback URLs**
   - Add `https://*.chromiumapp.org/` to Auth0 settings
   - Save the settings in Auth0 dashboard

4. **❌ Environment variables not loaded**
   - Make sure `.env.local` exists
   - Rebuild after changing environment variables
   - Variables must start with `VITE_`

### Check Current Configuration

Open the extension popup and check the console (F12) for debug logs that show your current configuration.

### Test Auth0 Configuration

You can test your Auth0 configuration by visiting:
```
https://YOUR-DOMAIN.auth0.com/authorize?client_id=YOUR-CLIENT-ID&response_type=code&redirect_uri=https://example.com/callback&scope=openid%20profile%20email
```

Replace `YOUR-DOMAIN` and `YOUR-CLIENT-ID` with your actual values.

## 📋 Checklist

- [ ] Auth0 application created as Single Page Application
- [ ] Callback URLs configured with `https://*.chromiumapp.org/`
- [ ] Domain and Client ID copied to `.env.local`
- [ ] Extension built with `npm run build`
- [ ] Extension loaded in Chrome developer mode

## 🆘 Still Need Help?

1. Check the browser console for detailed error messages
2. Verify your Auth0 application type is "Single Page Application"
3. Make sure your Auth0 domain is accessible (visit https://your-domain.auth0.com)
4. Try creating a fresh Auth0 application with the exact settings above 