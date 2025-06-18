'use client';

import React, { useState } from 'react';
import { ECDHService } from '../services/ecdh';

interface Message {
  id: string;
  sender: string;
  encryptedContent: string;
  iv: string;
  timestamp: number;
}

interface UserState {
  userId: string;
  keys: any;
  sharedAESKey: CryptoKey | null;
  sessionId: string;
  connected: boolean;
  loading: boolean;
  message: string;
}

export default function BackendECDH() {
  const [ecdhService] = useState(() => new ECDHService('/api'));
  const [messages, setMessages] = useState<Message[]>([]);
  
  // Separate state for Alice and Bob
  const [alice, setAlice] = useState<UserState>({
    userId: 'alice',
    keys: null,
    sharedAESKey: null,
    sessionId: '',
    connected: false,
    loading: false,
    message: ''
  });

  const [bob, setBob] = useState<UserState>({
    userId: 'bob',
    keys: null,
    sharedAESKey: null,
    sessionId: '',
    connected: false,
    loading: false,
    message: ''
  });

  const [logs, setLogs] = useState<string[]>([]);

  const addLog = (message: string) => {
    setLogs(prev => [...prev, `${new Date().toLocaleTimeString()}: ${message}`]);
  };

  // Connect user to backend
  const connectUser = async (userType: 'alice' | 'bob') => {
    const currentUser = userType === 'alice' ? alice : bob;
    const setUser = userType === 'alice' ? setAlice : setBob;
    const otherUserType = userType === 'alice' ? 'bob' : 'alice';

    setUser(prev => ({ ...prev, loading: true }));
    
    try {
      addLog(`🔄 ${userType.toUpperCase()} connecting to ${otherUserType.toUpperCase()}...`);

      // Use existing key pair if available, otherwise generate new one
      let myKeyPair = currentUser.keys;
      let myPublicKey;
      
      if (myKeyPair) {
        // Reuse existing key pair
        myPublicKey = await ecdhService.exportPublicKey(myKeyPair.publicKey);
        addLog(`🔑 ${userType.toUpperCase()} reusing existing key pair: ${myPublicKey.substring(0, 16)}...`);
      } else {
        // Generate new key pair only if we don't have one
        myKeyPair = await ecdhService.generateKeyPair();
        myPublicKey = await ecdhService.exportPublicKey(myKeyPair.publicKey);
        addLog(`🔑 ${userType.toUpperCase()} generated NEW key pair: ${myPublicKey.substring(0, 16)}...`);
        
        // Store the key pair immediately
        setUser(prev => ({ ...prev, keys: myKeyPair }));
      }
      
      // Connect to backend for peer-to-peer key exchange
      const response = await fetch('http://localhost:8080/api/key-exchange', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          myUserId: userType,
          otherUserId: otherUserType,
          myPublicKey: myPublicKey
        })
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const backendResponse = await response.json();
      
      if (backendResponse.success && backendResponse.theirPublicKey) {
        // Import other user's public key and derive AES key
        addLog(`🔑 ${userType.toUpperCase()} received ${otherUserType} public key: ${backendResponse.theirPublicKey.substring(0, 16)}...`);
        
        const otherUserPublicKey = await ecdhService.importPublicKey(backendResponse.theirPublicKey);
        const sharedSecret = await ecdhService.deriveSharedSecret(myKeyPair.privateKey, otherUserPublicKey);
        const aesKey = await ecdhService.deriveAESKey(sharedSecret);
        
        // Convert shared secret to hex for logging
        const sharedSecretHex = Array.from(new Uint8Array(sharedSecret))
          .map(b => b.toString(16).padStart(2, '0'))
          .join('');
        
        setUser(prev => ({
          ...prev,
          keys: myKeyPair,
          sharedAESKey: aesKey,
          sessionId: backendResponse.sessionId,
          connected: true,
          loading: false
        }));
        
        addLog(`✅ ${userType.toUpperCase()} connected with ${otherUserType.toUpperCase()}! Session: ${backendResponse.sessionId}`);
        addLog(`🔑 ${userType.toUpperCase()} shared secret: ${sharedSecretHex.substring(0, 32)}...`);
      } else {
        // Other user hasn't connected yet - don't retry, just wait for manual retry
        addLog(`⏳ ${userType.toUpperCase()} waiting for ${otherUserType.toUpperCase()} to connect...`);
        addLog(`💡 Click connect again once ${otherUserType} is ready`);
        setUser(prev => ({ ...prev, loading: false }));
      }
      
    } catch (error) {
      addLog(`❌ ${userType.toUpperCase()} connection failed: ${error}`);
      setUser(prev => ({ ...prev, loading: false }));
    }
  };

  // Send message from one user
  const sendMessage = async (senderType: 'alice' | 'bob') => {
    const sender = senderType === 'alice' ? alice : bob;
    const setSender = senderType === 'alice' ? setAlice : setBob;
    const receiver = senderType === 'alice' ? bob : alice;

    if (!sender.sharedAESKey || !sender.message || !receiver.connected) return;
    
    try {
      addLog(`📝 ${senderType.toUpperCase()} encrypting message...`);

      // Encrypt message
      const encryptedMessage = await ecdhService.encryptMessage(sender.message, sender.sharedAESKey);
      
      // Create message object
      const messageObj: Message = {
        id: 'msg_' + Date.now() + '_' + Math.random(),
        sender: senderType,
        encryptedContent: encryptedMessage.encryptedMessage,
        iv: encryptedMessage.iv,
        timestamp: encryptedMessage.timestamp
      };
      
      // Add message to the shared conversation
      setMessages(prev => [...prev, messageObj]);
      setSender(prev => ({ ...prev, message: '' }));
      
      addLog(`📤 ${senderType.toUpperCase()} sent encrypted message`);
      
      // Send to backend (optional - for persistence/relaying if needed)
      try {
        await fetch('http://localhost:8080/api/messages', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sessionId: sender.sessionId,
            encryptedData: encryptedMessage.encryptedMessage,
            iv: encryptedMessage.iv,
            timestamp: encryptedMessage.timestamp
          })
        });
      } catch (backendError) {
        // Backend storage is optional for this demo
        console.log('Backend storage failed (non-critical):', backendError);
      }
      
    } catch (error) {
      addLog(`❌ ${senderType.toUpperCase()} failed to send: ${error}`);
    }
  };

  // Fetch messages for a session (removed - not needed since we show messages locally)
  const fetchMessages = async (sessionId: string) => {
    // Removed automatic fetching to prevent duplicates
    // Messages are shown in real-time as they are sent
  };

  // Decrypt message for display
  const decryptMessage = async (msg: Message, userType: 'alice' | 'bob'): Promise<string> => {
    const user = userType === 'alice' ? alice : bob;
    
    if (!user.sharedAESKey) return '[No decryption key]';
    
    try {
      const decrypted = await ecdhService.decryptMessage({
        encryptedMessage: msg.encryptedContent,
        iv: msg.iv,
        timestamp: msg.timestamp
      }, user.sharedAESKey);
      return decrypted;
    } catch (error) {
      return '[Decryption failed]';
    }
  };

  const reset = () => {
    setAlice({
      userId: 'alice',
      keys: null,
      sharedAESKey: null,
      sessionId: '',
      connected: false,
      loading: false,
      message: ''
    });
    setBob({
      userId: 'bob',
      keys: null,
      sharedAESKey: null,
      sessionId: '',
      connected: false,
      loading: false,
      message: ''
    });
    setMessages([]);
    setLogs([]);
  };

  const resetBackend = async () => {
    try {
      await fetch('http://localhost:8080/api/reset', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      });
      addLog('🔄 Backend sessions reset');
      reset(); // Also reset frontend
    } catch (error) {
      addLog('❌ Failed to reset backend: ' + error);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-pink-100 p-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-4">
            🔐 Alice & Bob E2EE Chat Demo
          </h1>
          <p className="text-lg text-gray-600">
            True peer-to-peer E2EE: Each user has their own keys, backend only relays encrypted messages
          </p>
        </div>

        {/* Dual User Interface */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Alice's Side */}
          <UserChatBox
            user={alice}
            userType="alice"
            otherUser={bob}
            onConnect={() => connectUser('alice')}
            onSendMessage={() => sendMessage('alice')}
            onMessageChange={(message) => setAlice(prev => ({ ...prev, message }))}
            messages={messages}
            decryptMessage={decryptMessage}
          />

          {/* Bob's Side */}
          <UserChatBox
            user={bob}
            userType="bob"
            otherUser={alice}
            onConnect={() => connectUser('bob')}
            onSendMessage={() => sendMessage('bob')}
            onMessageChange={(message) => setBob(prev => ({ ...prev, message }))}
            messages={messages}
            decryptMessage={decryptMessage}
          />
        </div>

        {/* Connection Status */}
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <h3 className="text-lg font-semibold text-gray-800 mb-3">🔗 Connection Status</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className={`p-3 rounded-lg ${alice.connected ? 'bg-green-50 border-green-200' : 'bg-gray-50 border-gray-200'} border`}>
              <div className="flex items-center">
                <div className={`w-3 h-3 rounded-full mr-3 ${alice.connected ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                <span className="font-medium">👩 Alice</span>
              </div>
              {alice.connected && (
                <div className="text-xs text-gray-600 mt-1">Session: {alice.sessionId}</div>
              )}
            </div>
            <div className={`p-3 rounded-lg ${bob.connected ? 'bg-blue-50 border-blue-200' : 'bg-gray-50 border-gray-200'} border`}>
              <div className="flex items-center">
                <div className={`w-3 h-3 rounded-full mr-3 ${bob.connected ? 'bg-blue-500' : 'bg-gray-400'}`}></div>
                <span className="font-medium">👨 Bob</span>
              </div>
              {bob.connected && (
                <div className="text-xs text-gray-600 mt-1">Session: {bob.sessionId}</div>
              )}
            </div>
          </div>
        </div>

        {/* Process Logs */}
        {logs.length > 0 && (
          <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-800 mb-3">📋 Process Logs</h3>
            <div className="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm max-h-40 overflow-y-auto">
              {logs.map((log, index) => (
                <div key={index} className="mb-1">{log}</div>
              ))}
            </div>
          </div>
        )}

        {/* Controls */}
        <div className="flex justify-center gap-4">
          <button
            onClick={reset}
            className="bg-gray-600 hover:bg-gray-700 text-white font-medium py-2 px-6 rounded-lg transition-colors"
          >
            🔄 Reset Frontend
          </button>
          <button
            onClick={resetBackend}
            className="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-6 rounded-lg transition-colors"
          >
            🗑️ Reset Backend
          </button>
        </div>
      </div>
    </div>
  );
}

// Individual User Chat Component
function UserChatBox({ 
  user, 
  userType, 
  otherUser, 
  onConnect, 
  onSendMessage, 
  onMessageChange, 
  messages, 
  decryptMessage 
}: {
  user: UserState;
  userType: 'alice' | 'bob';
  otherUser: UserState;
  onConnect: () => void;
  onSendMessage: () => void;
  onMessageChange: (message: string) => void;
  messages: Message[];
  decryptMessage: (msg: Message, userType: 'alice' | 'bob') => Promise<string>;
}) {
  const userColor = userType === 'alice' ? 'green' : 'blue';
  const userName = userType === 'alice' ? '👩 Alice' : '👨 Bob';

  return (
    <div className={`bg-white rounded-lg shadow-lg border-2 border-${userColor}-200`}>
      <div className={`bg-${userColor}-50 p-4 rounded-t-lg border-b border-${userColor}-200`}>
        <h2 className={`text-xl font-semibold text-${userColor}-800`}>{userName}</h2>
      </div>
      
      <div className="p-6">
        {/* Connect Button */}
        {!user.connected ? (
          <div className="text-center mb-4">
            <button
              onClick={onConnect}
              disabled={user.loading}
              className={`bg-${userColor}-600 hover:bg-${userColor}-700 disabled:bg-${userColor}-300 text-white font-medium py-2 px-4 rounded-lg transition-colors`}
            >
              {user.loading ? '🔄 Connecting...' : '🚀 Connect to Backend'}
            </button>
          </div>
        ) : (
          <>
            {/* Chat Messages */}
            <div className="h-48 overflow-y-auto border border-gray-200 rounded-lg p-3 mb-4 bg-gray-50">
              {messages.map((msg) => (
                <MessageItem
                  key={msg.id}
                  message={msg}
                  userType={userType}
                  decryptMessage={decryptMessage}
                />
              ))}
            </div>

            {/* Message Input */}
            <div className="flex gap-2">
                <input
                type="text"
                value={user.message}
                onChange={(e) => onMessageChange(e.target.value)}
                placeholder={`Type message as ${userType}...`}
                className="flex-1 p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm text-black"
                onKeyPress={(e) => e.key === 'Enter' && onSendMessage()}
                disabled={!otherUser.connected}
                />
              <button
                onClick={onSendMessage}
                disabled={!user.message || !otherUser.connected}
                className={`bg-${userColor}-600 hover:bg-${userColor}-700 disabled:bg-gray-300 text-white font-medium px-4 py-2 rounded-lg transition-colors text-sm`}
              >
                🔒 Send
              </button>
            </div>

            {!otherUser.connected && (
              <p className="text-xs text-gray-500 mt-2 text-center">
                Waiting for {userType === 'alice' ? 'Bob' : 'Alice'} to connect...
              </p>
            )}
          </>
        )}
      </div>
    </div>
  );
}

// Message Display Component
function MessageItem({ 
  message, 
  userType, 
  decryptMessage 
}: {
  message: Message;
  userType: 'alice' | 'bob';
  decryptMessage: (msg: Message, userType: 'alice' | 'bob') => Promise<string>;
}) {
  const [decrypted, setDecrypted] = useState<string>('');
  const [loading, setLoading] = useState(true);

  React.useEffect(() => {
    decryptMessage(message, userType).then(result => {
      setDecrypted(result);
      setLoading(false);
    });
  }, [message, userType, decryptMessage]);

  const isOwn = message.sender === userType;
  const senderDisplay = message.sender === 'alice' ? '👩 Alice' : '👨 Bob';
  
  // Color scheme based on message sender
  const messageColor = message.sender === 'alice' 
    ? (isOwn ? 'bg-green-600 text-white' : 'bg-green-100 text-green-800')
    : (isOwn ? 'bg-blue-600 text-white' : 'bg-blue-100 text-blue-800');

  return (
    <div className={`mb-3 flex ${isOwn ? 'justify-end' : 'justify-start'}`}>
      <div className={`max-w-xs rounded-lg p-2 text-sm ${messageColor}`}>
        <div className="text-xs opacity-75 mb-1">
          {senderDisplay}
        </div>
        <div>
          {loading ? (
            <span className="text-xs opacity-75">🔓 Decrypting...</span>
          ) : (
            decrypted
          )}
        </div>
        <div className="text-xs opacity-75 mt-1">
          {new Date(message.timestamp).toLocaleTimeString()}
        </div>
      </div>
    </div>
  );
} 