'use client';

import React, { useState, useCallback } from 'react';
import { ECDHService, ECDHMessage, KeyPair } from '../services/ecdh';

interface KeyExchangeState {
  aliceKeys?: KeyPair;
  bobKeys?: KeyPair;
  sharedSecretAlice?: ArrayBuffer;
  sharedSecretBob?: ArrayBuffer;
  aesKeyAlice?: CryptoKey;
  aesKeyBob?: CryptoKey;
  alicePublicKeyDisplay?: string;
  bobPublicKeyDisplay?: string;
  sharedSecretHex?: string;
}

export default function ECDHDemo() {
  const [ecdhService] = useState(() => new ECDHService());
  const [keyExchangeState, setKeyExchangeState] = useState<KeyExchangeState>({});
  const [message, setMessage] = useState('');
  const [encryptedMessage, setEncryptedMessage] = useState<ECDHMessage | null>(null);
  const [decryptedMessage, setDecryptedMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [logs, setLogs] = useState<string[]>([]);

  const addLog = useCallback((message: string) => {
    setLogs(prev => [...prev, `${new Date().toLocaleTimeString()}: ${message}`]);
  }, []);

  const performKeyExchange = async () => {
    setLoading(true);
    setError('');
    setLogs([]);
    
    try {
      addLog('🔄 Starting ECDH key exchange...');
      
      const result = await ecdhService.simulateKeyExchange();
      
      addLog('🔑 Alice generated her key pair');
      addLog('🔑 Bob generated his key pair');
      addLog('🤝 Public keys exchanged');
      addLog('🔐 Shared secrets derived');
      addLog('🎯 AES keys derived from shared secrets');
      
      const alicePublicKeyDisplay = await ecdhService.formatPublicKeyForDisplay(result.aliceKeys.publicKey);
      const bobPublicKeyDisplay = await ecdhService.formatPublicKeyForDisplay(result.bobKeys.publicKey);
      const sharedSecretHex = ecdhService.arrayBufferToHex(result.sharedSecretAlice);
      
      setKeyExchangeState({
        ...result,
        alicePublicKeyDisplay,
        bobPublicKeyDisplay,
        sharedSecretHex: `${sharedSecretHex.substring(0, 32)}...`,
      });
      
      addLog('✅ Key exchange completed successfully!');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      addLog(`❌ Error: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  const encryptMessageFromAlice = async () => {
    if (!keyExchangeState.aesKeyAlice || !message) return;
    
    try {
      addLog(`📝 Alice encrypting message: "${message}"`);
      const encrypted = await ecdhService.encryptMessage(message, keyExchangeState.aesKeyAlice);
      setEncryptedMessage(encrypted);
      addLog('🔒 Message encrypted successfully');
      addLog(`📦 Encrypted data: ${encrypted.encryptedMessage.substring(0, 32)}...`);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      addLog(`❌ Encryption error: ${errorMessage}`);
    }
  };

  const decryptMessageForBob = async () => {
    if (!keyExchangeState.aesKeyBob || !encryptedMessage) return;
    
    try {
      addLog('🔓 Bob decrypting received message...');
      const decrypted = await ecdhService.decryptMessage(encryptedMessage, keyExchangeState.aesKeyBob);
      setDecryptedMessage(decrypted);
      addLog(`✅ Message decrypted: "${decrypted}"`);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      addLog(`❌ Decryption error: ${errorMessage}`);
    }
  };

  const reset = () => {
    setKeyExchangeState({});
    setMessage('');
    setEncryptedMessage(null);
    setDecryptedMessage('');
    setError('');
    setLogs([]);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-8">
      <div className="max-w-6xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-4">
            🔐 ECDH P-256 End-to-End Encryption Demo
          </h1>
          <p className="text-lg text-gray-600">
            Demonstrating Elliptic Curve Diffie-Hellman key exchange with AES-GCM encryption
          </p>
        </div>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-6">
            <strong>Error:</strong> {error}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Key Exchange Section */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4">🔑 Key Exchange</h2>
            
            <button
              onClick={performKeyExchange}
              disabled={loading}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white font-medium py-3 px-4 rounded-lg transition-colors mb-4"
            >
              {loading ? '🔄 Performing Key Exchange...' : '🚀 Start ECDH Key Exchange'}
            </button>

            {keyExchangeState.aliceKeys && (
              <div className="space-y-4">
                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                  <h3 className="font-medium text-green-800 mb-2">👩 Alice's Public Key</h3>
                  <code className="text-sm text-green-700 font-mono break-all">
                    {keyExchangeState.alicePublicKeyDisplay}
                  </code>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <h3 className="font-medium text-blue-800 mb-2">👨 Bob's Public Key</h3>
                  <code className="text-sm text-blue-700 font-mono break-all">
                    {keyExchangeState.bobPublicKeyDisplay}
                  </code>
                </div>

                <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
                  <h3 className="font-medium text-purple-800 mb-2">🔐 Shared Secret (Hex)</h3>
                  <code className="text-sm text-purple-700 font-mono break-all">
                    {keyExchangeState.sharedSecretHex}
                  </code>
                </div>
              </div>
            )}
          </div>

          {/* Message Encryption Section */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4">💬 Message Encryption</h2>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  📝 Message to Encrypt (Alice → Bob)
                </label>
                <textarea
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder="Enter your secret message here..."
                  className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  rows={3}
                />
              </div>

              <button
                onClick={encryptMessageFromAlice}
                disabled={!keyExchangeState.aesKeyAlice || !message}
                className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-300 text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                🔒 Encrypt Message (Alice)
              </button>

              {encryptedMessage && (
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <h3 className="font-medium text-yellow-800 mb-2">📦 Encrypted Message</h3>
                  <div className="text-sm text-yellow-700 font-mono space-y-1">
                    <div><strong>Data:</strong> {encryptedMessage.encryptedMessage.substring(0, 40)}...</div>
                    <div><strong>IV:</strong> {encryptedMessage.iv}</div>
                    <div><strong>Timestamp:</strong> {new Date(encryptedMessage.timestamp).toLocaleString()}</div>
                  </div>
                </div>
              )}

              <button
                onClick={decryptMessageForBob}
                disabled={!keyExchangeState.aesKeyBob || !encryptedMessage}
                className="w-full bg-purple-600 hover:bg-purple-700 disabled:bg-gray-300 text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                🔓 Decrypt Message (Bob)
              </button>

              {decryptedMessage && (
                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                  <h3 className="font-medium text-green-800 mb-2">✅ Decrypted Message</h3>
                  <div className="text-green-700 font-medium">"{decryptedMessage}"</div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Logs Section */}
        {logs.length > 0 && (
          <div className="mt-8 bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-2xl font-semibold text-gray-800 mb-4">📋 Process Logs</h2>
            <div className="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm max-h-60 overflow-y-auto">
              {logs.map((log, index) => (
                <div key={index} className="mb-1">
                  {log}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Reset and Info */}
        <div className="mt-8 flex flex-col sm:flex-row gap-4 items-center justify-between">
          <button
            onClick={reset}
            className="bg-gray-600 hover:bg-gray-700 text-white font-medium py-2 px-6 rounded-lg transition-colors"
          >
            🔄 Reset Demo
          </button>
          
          <div className="text-sm text-gray-600 text-center sm:text-right">
            <p>🔐 Using ECDH P-256 + AES-GCM-256</p>
            <p>🛡️ Perfect Forward Secrecy enabled</p>
          </div>
        </div>
      </div>
    </div>
  );
} 