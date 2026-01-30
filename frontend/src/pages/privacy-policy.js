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

      <h2>About this project</h2>
      <p>
        BakeFlow is a school project that lets customers place cake and bakery orders through Facebook Messenger. This
        page explains, in simple language, what information we see, what we store, how we use it, and what choices you
        have.
      </p>

      <h2>What data we collect</h2>
      <p>When you use our Facebook Page or Messenger experience, we may receive and store:</p>
      <ul>
        <li>
          Messages you send to our Page in Messenger, including text, emojis, stickers, images, and other content you
          choose to send.
        </li>
        <li>
          A page-scoped ID that Facebook generates for you so our system can know which messages and orders belong to
          the same person. We do not see your login password.
        </li>
        <li>
          Order details you provide, such as items, quantities, custom notes, delivery or pickup choice, and address (if
          you choose delivery).
        </li>
        <li>
          Contact details you share with us, such as your name and phone number, so we can confirm and manage your
          order.
        </li>
        <li>
          Optional information you give in follow-up chats, such as feedback, ratings, or special preferences.
        </li>
      </ul>

      <h2>Facebook profile information</h2>
      <p>
        To make the admin view easier to use, our system may briefly look up your Facebook display name and profile
        picture using Facebook&apos;s tools when an admin is viewing orders. This is only used to help the admin
        recognise who they are chatting with.
      </p>
      <p>
        We do not permanently store your full Facebook profile details in our database. The main identifier we store is
        the page-scoped user ID that Facebook provides for messaging.
      </p>

      <h2>How we use your information</h2>
      <ul>
        <li>To read and reply to your messages and questions in Messenger.</li>
        <li>To create, manage, and update your orders (for pickup or delivery).</li>
        <li>To send you order confirmations, status updates, and reminders.</li>
        <li>To help us understand common questions and improve the ordering experience.</li>
        <li>To keep simple records of orders for school project evaluation and basic business tracking.</li>
      </ul>

      <h2>How we store and protect data</h2>
      <p>
        We store your order information and Messenger conversations in a database so the system can function and admins
        can manage orders. We aim to keep this data as limited as possible and only for purposes described on this page.
      </p>
      <p>
        We take reasonable steps to protect your information from unauthorized access or misuse. However, because this
        is a school project and not a commercial platform, you should avoid sharing highly sensitive personal
        information through Messenger (such as passwords or government ID numbers).
      </p>

      <h2>Data retention</h2>
      <p>
        We keep messages and order data for as long as needed to operate the service, review the project, and maintain
        basic records. Data may be kept for a period after your order is completed so that admins can see past orders
        and so we can evaluate how the project performed.
      </p>
      <p>
        If you request deletion of your data (see below), we will do our best to remove or anonymise personally
        identifiable information from our database, while still keeping any records we must keep for technical,
        reporting, or legal reasons.
      </p>

      <h2>Sharing of information</h2>
      <p>
        We do not sell your personal information. We may share your information only in the following situations:
      </p>
      <ul>
        <li>
          With service providers we use to run this project or deliver your order (for example, tools for hosting,
          storing data, or delivering messages).
        </li>
        <li>With teachers, mentors, or evaluators who are reviewing this project for school purposes.</li>
        <li>When required by law, regulation, or a valid legal request.</li>
      </ul>

      <h2>Your choices and data deletion</h2>
      <p>You have control over the information you share with us:</p>
      <ul>
        <li>You can choose what to type or send in Messenger and what contact details to provide.</li>
        <li>You can stop using the service at any time by no longer messaging the Page.</li>
        <li>
          You can request to see what information we have about your orders and Messenger conversations linked to your
          account.
        </li>
        <li>
          You can request deletion of your data, where possible, and we will delete or anonymise the information we are
          not required to keep.
        </li>
      </ul>
      <p>
        To make a request, you can message our Facebook Page or email{" "}
        <strong>rinnyaluvu@gmail.com</strong> and clearly state that you want to access or delete your data.
      </p>

      <h2>Third-party platforms</h2>
      <p>
        This project relies on Facebook and Messenger. Your use of our service is also covered by Facebook&apos;s own
        terms and policies, including their Data Policy. We do not control how Facebook itself collects or uses data
        when you use their apps or websites.
      </p>

      <h2>Children&apos;s privacy</h2>
      <p>
        This project is designed for general customers and is not specifically directed at children. If you believe we
        have collected personal information from a child in a way that concerns you, please contact us and we will
        review and address the situation.
      </p>

      <h2>Changes to this policy</h2>
      <p>
        Because this is a school project, the system may change or be improved over time. If we make material changes to
        how we handle your information, we will update this page and adjust the &quot;Last updated&quot; date at the
        top.
      </p>

      <h2>Contact</h2>
      <p>
        If you have any questions or concerns about this Privacy Policy or how your data is handled, you can contact:
      </p>
      <p>
        Email: <strong>rinnyaluvu@gmail.com</strong>
        <br />
        Project name: <strong>BakeFlow (School Project)</strong>
      </p>
    </main>
  );
}
