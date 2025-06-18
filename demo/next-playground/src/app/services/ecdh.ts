export interface KeyPair {
  publicKey: CryptoKey;
  privateKey: CryptoKey;
}

export interface ECDHMessage {
  encryptedMessage: string;
  iv: string;
  timestamp: number;
}

export class ECDHService {
  private apiEndpoint: string;

  constructor(apiEndpoint: string = '/api') {
    this.apiEndpoint = apiEndpoint;
  }

  // Generate ephemeral ECDH key pair
  async generateKeyPair(): Promise<KeyPair> {
    try {
      const keyPair = await crypto.subtle.generateKey(
        {
          name: "ECDH",
          namedCurve: "P-256",
        },
        true,
        ["deriveKey", "deriveBits"]
      );
      return keyPair as KeyPair;
    } catch (error) {
      throw new Error(`Failed to generate key pair: ${error}`);
    }
  }

  // Export public key for exchange
  async exportPublicKey(publicKey: CryptoKey): Promise<string> {
    try {
      const exported = await crypto.subtle.exportKey("raw", publicKey);
      return btoa(String.fromCharCode(...new Uint8Array(exported)));
    } catch (error) {
      throw new Error(`Failed to export public key: ${error}`);
    }
  }

  // Import public key from base64
  async importPublicKey(publicKeyBase64: string): Promise<CryptoKey> {
    try {
      const keyData = Uint8Array.from(atob(publicKeyBase64), c => c.charCodeAt(0));
      return await crypto.subtle.importKey(
        "raw",
        keyData,
        {
          name: "ECDH",
          namedCurve: "P-256",
        },
        false,
        []
      );
    } catch (error) {
      throw new Error(`Failed to import public key: ${error}`);
    }
  }

  // Derive shared secret between two keys
  async deriveSharedSecret(privateKey: CryptoKey, publicKey: CryptoKey): Promise<ArrayBuffer> {
    try {
      return await crypto.subtle.deriveBits(
        {
          name: "ECDH",
          public: publicKey,
        },
        privateKey,
        256
      );
    } catch (error) {
      throw new Error(`Failed to derive shared secret: ${error}`);
    }
  }

  // Derive AES key from shared secret
  async deriveAESKey(sharedSecret: ArrayBuffer): Promise<CryptoKey> {
    try {
      const keyMaterial = await crypto.subtle.importKey(
        "raw",
        sharedSecret,
        { name: "HKDF" },
        false,
        ["deriveKey"]
      );

      return await crypto.subtle.deriveKey(
        {
          name: "HKDF",
          hash: "SHA-256",
          salt: new Uint8Array(),
          info: new TextEncoder().encode("ECDH-AES-GCM"),
        },
        keyMaterial,
        { name: "AES-GCM", length: 256 },
        false,
        ["encrypt", "decrypt"]
      );
    } catch (error) {
      throw new Error(`Failed to derive AES key: ${error}`);
    }
  }

  // Encrypt message using derived AES key
  async encryptMessage(message: string, aesKey: CryptoKey): Promise<ECDHMessage> {
    try {
      const iv = crypto.getRandomValues(new Uint8Array(12));
      const encodedMessage = new TextEncoder().encode(message);

      const encryptedData = await crypto.subtle.encrypt(
        {
          name: "AES-GCM",
          iv: iv,
        },
        aesKey,
        encodedMessage
      );

      return {
        encryptedMessage: btoa(String.fromCharCode(...new Uint8Array(encryptedData))),
        iv: btoa(String.fromCharCode(...iv)),
        timestamp: Date.now(),
      };
    } catch (error) {
      throw new Error(`Failed to encrypt message: ${error}`);
    }
  }

  // Decrypt message using derived AES key
  async decryptMessage(ecdhMessage: ECDHMessage, aesKey: CryptoKey): Promise<string> {
    try {
      const encryptedData = Uint8Array.from(atob(ecdhMessage.encryptedMessage), c => c.charCodeAt(0));
      const iv = Uint8Array.from(atob(ecdhMessage.iv), c => c.charCodeAt(0));

      const decryptedData = await crypto.subtle.decrypt(
        {
          name: "AES-GCM",
          iv: iv,
        },
        aesKey,
        encryptedData
      );

      return new TextDecoder().decode(decryptedData);
    } catch (error) {
      throw new Error(`Failed to decrypt message: ${error}`);
    }
  }

  // Complete key exchange process (for demo purposes - simulates two parties)
  async simulateKeyExchange(): Promise<{
    aliceKeys: KeyPair;
    bobKeys: KeyPair;
    sharedSecretAlice: ArrayBuffer;
    sharedSecretBob: ArrayBuffer;
    aesKeyAlice: CryptoKey;
    aesKeyBob: CryptoKey;
  }> {
    try {
      // Generate key pairs for both parties
      const aliceKeys = await this.generateKeyPair();
      const bobKeys = await this.generateKeyPair();

      // Each party derives the shared secret using their private key and the other's public key
      const sharedSecretAlice = await this.deriveSharedSecret(aliceKeys.privateKey, bobKeys.publicKey);
      const sharedSecretBob = await this.deriveSharedSecret(bobKeys.privateKey, aliceKeys.publicKey);

      // Both secrets should be identical
      const aliceSecretArray = new Uint8Array(sharedSecretAlice);
      const bobSecretArray = new Uint8Array(sharedSecretBob);
      
      if (aliceSecretArray.length !== bobSecretArray.length || 
          !aliceSecretArray.every((val, i) => val === bobSecretArray[i])) {
        throw new Error("Shared secrets do not match!");
      }

      // Derive AES keys from shared secrets
      const aesKeyAlice = await this.deriveAESKey(sharedSecretAlice);
      const aesKeyBob = await this.deriveAESKey(sharedSecretBob);

      return {
        aliceKeys,
        bobKeys,
        sharedSecretAlice,
        sharedSecretBob,
        aesKeyAlice,
        aesKeyBob,
      };
    } catch (error) {
      throw new Error(`Key exchange simulation failed: ${error}`);
    }
  }

  // Helper function to convert ArrayBuffer to hex string for display
  arrayBufferToHex(buffer: ArrayBuffer): string {
    return Array.from(new Uint8Array(buffer))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  }

  // Helper function to format public key for display
  async formatPublicKeyForDisplay(publicKey: CryptoKey): Promise<string> {
    try {
      const exported = await this.exportPublicKey(publicKey);
      return `${exported.substring(0, 20)}...${exported.substring(exported.length - 20)}`;
    } catch (error) {
      return "Error formatting key";
    }
  }
} 