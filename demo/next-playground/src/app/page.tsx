'use client';

import { useState } from 'react';
import ECDHDemo from './components/ECDHDemo';
import BackendECDH from './components/BackendECDH';

export default function Home() {
  const [activeTab, setActiveTab] = useState<'backend' | 'demo'>('backend');

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Navigation */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-6xl mx-auto px-8">
          <div className="flex space-x-8">
            <button
              onClick={() => setActiveTab('backend')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === 'backend'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              🌐 Backend Integration
            </button>
            <button
              onClick={() => setActiveTab('demo')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === 'demo'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              🔐 ECDH Demo
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="w-full">
        {activeTab === 'backend' && <BackendECDH />}
        {activeTab === 'demo' && <ECDHDemo />}
      </div>
    </div>
  );
}
