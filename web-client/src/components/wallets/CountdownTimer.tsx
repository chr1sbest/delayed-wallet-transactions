'use client';

import { useState, useEffect } from 'react';

interface CountdownTimerProps {
  executionTime: Date;
}

export function CountdownTimer({ executionTime }: CountdownTimerProps) {
  const [remainingTime, setRemainingTime] = useState('');
  const [isExpired, setIsExpired] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      const now = new Date();
      const timeLeft = executionTime.getTime() - now.getTime();

      if (timeLeft <= 0) {
        setRemainingTime('0:00');
        setIsExpired(true);
        clearInterval(interval);
        return;
      }

      const minutes = Math.floor((timeLeft / 1000) / 60);
      const seconds = Math.floor((timeLeft / 1000) % 60);

      setRemainingTime(`${minutes}:${seconds.toString().padStart(2, '0')}`);
    }, 1000);

    return () => clearInterval(interval);
  }, [executionTime]);

  if (!remainingTime) {
    return null;
  }

  return (
    <span className={`px-2 py-1 text-xs font-medium rounded-full ${isExpired ? 'bg-red-500 text-white' : 'bg-gray-200 text-gray-800'}`}>
      {remainingTime}
    </span>
  );
}
