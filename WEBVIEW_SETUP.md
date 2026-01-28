# How to Enable Per-Product Ratings with Webview in Messenger

Follow these steps to let customers rate each product with clickable stars in Messenger:

## 1. Install ngrok (if not already installed)

- Download from https://ngrok.com/download
- Or install via Homebrew:
  ```sh
  brew install ngrok/ngrok/ngrok
  ```

## 2. Start your frontend server

- In a terminal, go to your frontend folder:
  ```sh
  cd /Users/zuuji/Desktop/BakeFlow/frontend
  npm run dev
  ```
  (or however you normally start your frontend)

## 3. Expose your frontend with ngrok

- In a new terminal:
  ```sh
  ngrok http 3000
  ```
- Copy the HTTPS URL shown (e.g. `https://abc123.ngrok.io`)

## 4. Set WEBVIEW_BASE_URL in your backend .env file

- Open `/Users/zuuji/Desktop/BakeFlow/backend/.env`
- Add or update this line:
  ```
  WEBVIEW_BASE_URL=https://abc123.ngrok.io
  ```
  (replace with your actual ngrok HTTPS URL)

## 5. Restart your backend server

- In your backend terminal:
  ```sh
  go run main.go
  ```

## 6. Test the flow

- Mark an order as delivered in the admin panel
- The customer will get a Messenger card with a “⭐ Rate Order” button
- Tapping it opens the webview with clickable stars for each product

---

**Note:**
- The webview only works if Messenger can access your frontend via HTTPS (ngrok or real domain)
- You can customize the webview UI in `frontend/public/rate-order.html`

If you need help with any step, just ask!
