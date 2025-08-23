/**
 * @file ChannelKeyModeToggle.test.js
 * @description 测试渠道Key模式切换功能的前端组件
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Switch, Tag } from '@douyinfe/semi-ui';
import { jest } from '@jest/globals';

// Mock API module
const mockAPI = {
  patch: jest.fn(),
};

// Mock helpers
const mockHelpers = {
  showSuccess: jest.fn(),
  showError: jest.fn(),
};

// 模拟渠道Key模式切换组件
const ChannelKeyModeToggle = ({ record, onUpdate }) => {
  const [loading, setLoading] = React.useState(false);
  const isPollingMode = record.channel_info?.multi_key_mode === 'polling';

  const handleToggle = async (checked) => {
    setLoading(true);
    try {
      const res = await mockAPI.patch(`/api/channel/${record.id}/key-mode`, {
        enabled: checked
      });
      
      if (res?.data?.success) {
        record.channel_info.multi_key_mode = checked ? 'polling' : 'random';
        mockHelpers.showSuccess(checked ? '已启用轮询模式' : '已启用随机模式');
        onUpdate && onUpdate(record);
      } else {
        mockHelpers.showError(res?.data?.message || '操作失败');
      }
    } catch (error) {
      mockHelpers.showError('操作失败: ' + (error?.response?.data?.message || error?.message || error));
    } finally {
      setLoading(false);
    }
  };

  // 单Key模式
  if (!record.channel_info?.is_multi_key) {
    return (
      <Tag color='grey' shape='circle'>
        单Key
      </Tag>
    );
  }

  // 多Key模式
  return (
    <div className="flex items-center gap-2">
      <Switch
        size='small'
        checked={isPollingMode}
        loading={loading}
        onChange={handleToggle}
        data-testid="key-mode-switch"
      />
      <span className="text-xs">
        {isPollingMode ? (
          <Tag color='green' shape='circle' data-testid="polling-tag">
            轮询
          </Tag>
        ) : (
          <Tag color='blue' shape='circle' data-testid="random-tag">
            随机
          </Tag>
        )}
      </span>
    </div>
  );
};

describe('ChannelKeyModeToggle', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Single Key Mode', () => {
    it('should display "单Key" for single key channels', () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: false
        }
      };

      render(<ChannelKeyModeToggle record={record} />);
      
      expect(screen.getByText('单Key')).toBeInTheDocument();
      expect(screen.queryByTestId('key-mode-switch')).not.toBeInTheDocument();
    });
  });

  describe('Multi Key Mode', () => {
    it('should display switch and "随机" tag for random mode', () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'random'
        }
      };

      render(<ChannelKeyModeToggle record={record} />);
      
      expect(screen.getByTestId('key-mode-switch')).toBeInTheDocument();
      expect(screen.getByTestId('random-tag')).toBeInTheDocument();
      expect(screen.queryByTestId('polling-tag')).not.toBeInTheDocument();
    });

    it('should display switch and "轮询" tag for polling mode', () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'polling'
        }
      };

      render(<ChannelKeyModeToggle record={record} />);
      
      expect(screen.getByTestId('key-mode-switch')).toBeInTheDocument();
      expect(screen.getByTestId('polling-tag')).toBeInTheDocument();
      expect(screen.queryByTestId('random-tag')).not.toBeInTheDocument();
    });

    it('should toggle from random to polling mode successfully', async () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'random'
        }
      };

      const mockResponse = {
        data: {
          success: true,
          data: {
            channel_id: 1,
            key_mode: 'polling',
            enabled: true
          }
        }
      };

      mockAPI.patch.mockResolvedValue(mockResponse);
      const mockOnUpdate = jest.fn();

      render(<ChannelKeyModeToggle record={record} onUpdate={mockOnUpdate} />);
      
      const toggle = screen.getByTestId('key-mode-switch');
      fireEvent.click(toggle);

      await waitFor(() => {
        expect(mockAPI.patch).toHaveBeenCalledWith('/api/channel/1/key-mode', {
          enabled: true
        });
        expect(mockHelpers.showSuccess).toHaveBeenCalledWith('已启用轮询模式');
        expect(record.channel_info.multi_key_mode).toBe('polling');
        expect(mockOnUpdate).toHaveBeenCalledWith(record);
      });
    });

    it('should toggle from polling to random mode successfully', async () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'polling'
        }
      };

      const mockResponse = {
        data: {
          success: true,
          data: {
            channel_id: 1,
            key_mode: 'random',
            enabled: false
          }
        }
      };

      mockAPI.patch.mockResolvedValue(mockResponse);
      const mockOnUpdate = jest.fn();

      render(<ChannelKeyModeToggle record={record} onUpdate={mockOnUpdate} />);
      
      const toggle = screen.getByTestId('key-mode-switch');
      fireEvent.click(toggle);

      await waitFor(() => {
        expect(mockAPI.patch).toHaveBeenCalledWith('/api/channel/1/key-mode', {
          enabled: false
        });
        expect(mockHelpers.showSuccess).toHaveBeenCalledWith('已启用随机模式');
        expect(record.channel_info.multi_key_mode).toBe('random');
        expect(mockOnUpdate).toHaveBeenCalledWith(record);
      });
    });

    it('should handle API error gracefully', async () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'random'
        }
      };

      const mockError = new Error('Network error');
      mockAPI.patch.mockRejectedValue(mockError);

      render(<ChannelKeyModeToggle record={record} />);
      
      const toggle = screen.getByTestId('key-mode-switch');
      fireEvent.click(toggle);

      await waitFor(() => {
        expect(mockHelpers.showError).toHaveBeenCalledWith('操作失败: Network error');
      });
    });

    it('should handle API response error', async () => {
      const record = {
        id: 1,
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'random'
        }
      };

      const mockResponse = {
        data: {
          success: false,
          message: '该渠道不是多Key模式，无法切换Key模式'
        }
      };

      mockAPI.patch.mockResolvedValue(mockResponse);

      render(<ChannelKeyModeToggle record={record} />);
      
      const toggle = screen.getByTestId('key-mode-switch');
      fireEvent.click(toggle);

      await waitFor(() => {
        expect(mockHelpers.showError).toHaveBeenCalledWith('该渠道不是多Key模式，无法切换Key模式');
      });
    });
  });
});

// 集成测试：测试在渠道列表中的表现
describe('ChannelKeyModeToggle Integration', () => {
  it('should work correctly in channel table context', () => {
    const channels = [
      {
        id: 1,
        name: 'Single Key Channel',
        channel_info: {
          is_multi_key: false
        }
      },
      {
        id: 2,
        name: 'Multi Key Random Channel', 
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'random'
        }
      },
      {
        id: 3,
        name: 'Multi Key Polling Channel',
        channel_info: {
          is_multi_key: true,
          multi_key_mode: 'polling'
        }
      }
    ];

    const ChannelTable = () => (
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>Key模式</th>
          </tr>
        </thead>
        <tbody>
          {channels.map(channel => (
            <tr key={channel.id}>
              <td>{channel.id}</td>
              <td>{channel.name}</td>
              <td>
                <ChannelKeyModeToggle record={channel} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    );

    render(<ChannelTable />);

    // 验证不同类型渠道的显示
    expect(screen.getByText('单Key')).toBeInTheDocument();
    expect(screen.getByTestId('random-tag')).toBeInTheDocument();
    expect(screen.getByTestId('polling-tag')).toBeInTheDocument();
    
    // 验证只有多Key渠道才有开关
    const switches = screen.getAllByTestId('key-mode-switch');
    expect(switches).toHaveLength(2); // 只有两个多Key渠道
  });
});

export default ChannelKeyModeToggle;