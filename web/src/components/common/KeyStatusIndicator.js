import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Space,
  Spin,
  Tag,
  Tooltip,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconCheckCircle,
  IconAlertTriangle,
  IconXCircle,
  IconPulse,
  IconStop,
} from '@douyinfe/semi-icons';
import { API, showError } from '../../helpers';

const { Text } = Typography;

const KeyStatusIndicator = ({ channelId, refreshTrigger }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [keyStatus, setKeyStatus] = useState(null);

  useEffect(() => {
    if (channelId) {
      fetchKeyStatus();
    }
  }, [channelId, refreshTrigger]);

  const fetchKeyStatus = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/channel/${channelId}/key?view_mode=count`);
      if (res.data.success) {
        setKeyStatus(res.data.data);
      } else {
        console.error('Failed to fetch key status:', res.data.message);
        setKeyStatus(null);
      }
    } catch (error) {
      console.error('Error fetching key status:', error);
      setKeyStatus(null);
    } finally {
      setLoading(false);
    }
  };

  const getStatusInfo = () => {
    if (!keyStatus) {
      return {
        icon: <IconStop style={{ color: '#d9d9d9' }} />,
        color: 'grey',
        text: t('未配置'),
        tooltip: t('该渠道未配置密钥'),
      };
    }

    if (keyStatus.key_count === 0) {
      return {
        icon: <IconStop style={{ color: '#d9d9d9' }} />,
        color: 'grey',
        text: t('未配置'),
        tooltip: t('该渠道未配置密钥'),
      };
    }

    // 检查是否有禁用的密钥
    if (keyStatus.keys && keyStatus.keys.length > 0) {
      const activeKeys = keyStatus.keys.filter(key => key.status === 'active');
      const disabledKeys = keyStatus.keys.filter(key => key.status === 'disabled');
      
      if (activeKeys.length === 0) {
        // 全部失效
        return {
          icon: <IconXCircle style={{ color: '#ff4d4f' }} />,
          color: 'red',
          text: t('全部失效'),
          tooltip: t('{{total}}个密钥全部不可用', { total: keyStatus.key_count }),
        };
      } else if (disabledKeys.length > 0) {
        // 部分失效
        return {
          icon: <IconAlertTriangle style={{ color: '#faad14' }} />,
          color: 'orange',
          text: t('部分失效'),
          tooltip: t('{{active}}个可用，{{disabled}}个失效', { 
            active: activeKeys.length, 
            disabled: disabledKeys.length 
          }),
        };
      }
    }

    // 全部正常
    if (keyStatus.key_count === 1) {
      return {
        icon: <IconCheckCircle style={{ color: '#52c41a' }} />,
        color: 'green',
        text: t('正常'),
        tooltip: t('1个密钥正常'),
      };
    } else {
      return {
        icon: <IconCheckCircle style={{ color: '#52c41a' }} />,
        color: 'green',
        text: t('全部正常'),
        tooltip: t('{{count}}个密钥全部正常', { count: keyStatus.key_count }),
      };
    }
  };

  if (loading) {
    return (
      <Space>
        <Spin size="small" />
        <Text type="tertiary" size="small">{t('检测中...')}</Text>
      </Space>
    );
  }

  const statusInfo = getStatusInfo();

  return (
    <Tooltip content={statusInfo.tooltip} position="top">
      <Tag 
        color={statusInfo.color} 
        size="small"
        style={{ 
          cursor: 'help',
          display: 'flex',
          alignItems: 'center',
        }}
      >
        {statusInfo.icon}
        <span style={{ marginLeft: 4 }}>
          {statusInfo.text}
          {keyStatus && keyStatus.key_count > 1 && (
            <span style={{ marginLeft: 4, opacity: 0.8 }}>
              ({keyStatus.key_count})
            </span>
          )}
        </span>
      </Tag>
    </Tooltip>
  );
};

export default KeyStatusIndicator;