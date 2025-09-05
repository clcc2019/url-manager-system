import { formatDate, formatDuration, getTimeUntilExpiry } from './date';
import { describe, it, expect } from 'vitest';

describe('Date Utils', () => {
  describe('formatDate', () => {
    it('formats date string correctly', () => {
      const dateString = '2023-12-25T10:30:45Z';
      const result = formatDate(dateString);
      
      // The exact format depends on locale, but should contain date components
      expect(result).toContain('2023');
      expect(result).toContain('25');
      expect(result).toContain('10');
      expect(result).toContain('30');
    });

    it('handles invalid date string', () => {
      const invalidDate = 'invalid-date';
      const result = formatDate(invalidDate);
      
      // Should return some string (Invalid Date or similar)
      expect(typeof result).toBe('string');
    });
  });

  describe('formatDuration', () => {
    it('formats seconds correctly', () => {
      expect(formatDuration(30)).toBe('0分钟');
      expect(formatDuration(90)).toBe('1分钟');
      expect(formatDuration(3600)).toBe('1小时0分钟');
      expect(formatDuration(3690)).toBe('1小时1分钟');
      expect(formatDuration(86400)).toBe('1天0小时');
      expect(formatDuration(90000)).toBe('1天1小时');
    });

    it('handles zero seconds', () => {
      expect(formatDuration(0)).toBe('0分钟');
    });

    it('handles large numbers', () => {
      const result = formatDuration(172800); // 2 days
      expect(result).toContain('2天');
    });
  });

  describe('getTimeUntilExpiry', () => {
    it('returns expired for past dates', () => {
      const pastDate = new Date(Date.now() - 1000).toISOString();
      expect(getTimeUntilExpiry(pastDate)).toBe('已过期');
    });

    it('returns formatted duration for future dates', () => {
      const futureDate = new Date(Date.now() + 3600000).toISOString(); // 1 hour from now
      const result = getTimeUntilExpiry(futureDate);
      
      expect(result).toContain('小时');
      expect(result).not.toBe('已过期');
    });

    it('handles very close future dates', () => {
      const closeDate = new Date(Date.now() + 30000).toISOString(); // 30 seconds from now
      const result = getTimeUntilExpiry(closeDate);
      
      expect(result).not.toBe('已过期');
    });
  });
});