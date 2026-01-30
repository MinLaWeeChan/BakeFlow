import React from "react";

export default function PrivacyPolicy() {
  return (
    <main
      style={{
        maxWidth: 900,
        margin: "0 auto",
        padding: "32px 16px",
        lineHeight: 1.65,
      }}
    >
      <h1>Privacy Policy</h1>
      <p>Last updated: {new Date().toISOString().slice(0, 10)}</p>

      <h2>Who we are</h2>
      <p>
        This Privacy Policy explains how we collect, use, and protect information when you interact with our Facebook
        Page and Messenger ordering experience.
      </p>

      <h2>Information we collect</h2>
      <ul>
        <li>Messages you send to our Page and related Messenger metadata (such as page-scoped user ID)</li>
        <li>Order information you provide (items, quantity, delivery/pickup details)</li>
        <li>Contact details you provide (name and phone number)</li>
      </ul>

      <h2>How we use your information</h2>
      <ul>
        <li>To reply to your messages and provide customer support</li>
        <li>To process and fulfill your orders (delivery or pickup)</li>
        <li>To send order updates (confirmation, status, changes)</li>
      </ul>

      <h2>How we store and protect data</h2>
      <p>
        We store information only as needed to operate the service and fulfill orders. We take reasonable measures to
        protect your data from unauthorized access.
      </p>

      <h2>Sharing</h2>
      <p>
        We do not sell your personal information. We may share information only with service providers needed to operate
        the service (for example, delivery coordination) or when required by law.
      </p>

      <h2>Data retention</h2>
      <p>
        We keep data only as long as needed for order processing, support, and basic business records, unless a longer
        retention period is required by law.
      </p>

      <h2>Your choices and data deletion</h2>
      <p>
        You can request access or deletion of your data by messaging our Page or emailing us at{" "}
        <strong>YOUR_EMAIL_HERE</strong>. If you message us “delete my data”, we will provide next steps.
      </p>

      <h2>Contact</h2>
      <p>
        Email: <strong>YOUR_EMAIL_HERE</strong>
        <br />
        Business name: <strong>YOUR_BUSINESS_NAME_HERE</strong>
      </p>
    </main>
  );
}
