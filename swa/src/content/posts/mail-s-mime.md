---
layout: '@layouts/BlogPost.astro'
title: 'Signing Emails with S/MIME'
description: 'A guide on how to sign your emails using S/MIME on macOS and iOS to ensure the authenticity and integrity of your emails.'
author: Clowa
draft: false
pubDate: 2026-01-04
tags: ["email", "security", "tutorial", "S/MIME"]
---

Email is the backbone of modern communication, but it’s also a prime target for phishing and impersonation. If you want to ensure your messages are authentic and tamper-proof, signing your emails with S/MIME is a must. In this guide, I’ll walk you through the process of setting up S/MIME for email signing on macOS and iOS, so you can protect your digital identity with confidence.

## What is S/MIME and Why Use It?

S/MIME (Secure/Multipurpose Internet Mail Extensions) is a standard for public key encryption and signing of email messages. When you sign an email with S/MIME, recipients can verify that the message really came from you and that it hasn’t been altered in transit. This is especially important for business communications, sensitive information, or simply peace of mind.

## Step 1: Obtain an S/MIME Certificate

Before you can sign emails, you need an S/MIME certificate from a trusted Certificate Authority (CA). Some providers offer free certificates, while others charge a fee. For this guide, I’ll use [Posteo](https://posteo.de/en) as an example, but the steps are similar for other providers.

1. **Request a Certificate:** Visit your provider’s website and follow their instructions to generate an S/MIME certificate. For Posteo users, check out their official [guide](https://posteo.de/en/help/how-do-i-create-an-smime-certificate).
2. **Store Your Certificate Securely:** Once you receive your certificate, save it in a secure location. A password manager like 1Password is ideal for storing both the certificate file and its password. If you prefer local storage, make sure the file is encrypted with a strong password—your private key must remain secret. The easiest format is a `.pfx` aka `.p12` file, which bundles your certificate and private key together.
  
>[!TIP]
> Never share your private key. If someone else gets access, they can impersonate you.

## Step 2: Set Up S/MIME on macOS

Getting S/MIME working on your Mac is straightforward:

1. **Install the Certificate:** Double-click the certificate file and follow the prompts to add it to your Keychain.
2. **Verify in Mail App:** Open the Mail app and compose a new email. If everything is set up correctly, you’ll see a checkmark icon in the compose window, indicating your email will be signed.

That’s it! Your outgoing emails from your macOS device are now signed, and recipients can verify their authenticity.

## Step 3: Set Up S/MIME on iOS

Want to sign emails from your iPhone or iPad? Here’s how:

1. **Transfer the Certificate:** Send the certificate file to your device (AirDrop, email, or password manager).
2. **Install the Profile:** Tap the file to start installation. You’ll see a “Profile Downloaded” message — go to Settings to finish the process.
3. **Complete Installation:** In the Settings App, at the top, confirm the new profile. Enter your device password and the certificate’s password when prompted.
4. **Configure Mail Settings:**
   - Go to Settings > Mail > Accounts > [Your Account] > Account > Advanced > S/MIME.
   - Enable `Sign` to sign outgoing emails by default. You can also enable `Encrypt` if you want to encrypt emails to recipients who have S/MIME.
5. **Verify in Mail App:** When composing a new email, you will find a text saying `Signed` or `Encrypted` at the very top of the compose window.

## Final Thoughts

S/MIME is a powerful option for anyone who values privacy and authenticity in email communication. By following these steps, you’ll be able to sign your emails on both macOS and iOS, giving your contacts confidence that your messages are genuine.

## Sources

- [How do I use end-to-end encryption with S/MIME in Apple Mail for macOS](https://posteo.de/en/help/how-do-i-use-end-to-end-encryption-with-smime-in-apple-mail-for-macOS)
- [Use S/MIME to send and receive encrypted messages in the Mail app in iOS](https://support.apple.com/en-us/102245)
- [How do I use end-to-end encryption with S/MIME on my iPhone or iPad?](https://posteo.de/en/help/how-do-i-use-end-to-end-encryption-with-smime-on-my-iphone-or-ipad)
- [S/MIME on IOS (IPhone, IPad) (DE)](https://www.kim.uni-konstanz.de/e-mail-und-internet/it-sicherheit/sichere-e-mail-kommunikation/schritt-3-e-mail-client-einrichten/s-mime-unter-ios-iphone-ipad/)
