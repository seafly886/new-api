import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Modal,
  Typography,
  Space,
  Button,
  Spin,
  List,
  Tag,
  Descriptions,
  message,
  Tooltip,
  Card,
  Divider,
} from '@douyinfe/semi-ui';
import {
  IconEye,
  IconClose,
  IconCopy,
  IconPulse,
  IconCheckCircleStroked,
  IconAlertTriangle,
  IconXCircle,
} from '@douyinfe/semi-icons';
import { API, showError, showSuccess, copy } from '../../helpers';

const { Text, Title } = Typography;

const KeyViewModal = ({ visible, onClose, channelId, channelName }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [keyData, setKeyData] = useState(null);

  useEffect(() => {
    if (visible && channelId) {
      fetchChannelKey();
    }
  }, [visible, channelId]);

  const fetchChannelKey = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/channel/${channelId}/key?view_mode=masked`);
      if (res.data.success) {
        setKeyData(res.data.data);
      } else {
        showError(res.data.message || t('获取密钥信息失败'));
      }
    } catch (error) {
      showError(t('获取密钥信息失败: {{msg}}', { msg: error.message }));
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = (maskedKey, index) => {
    // 注意：这里只能复制脱敏后的密钥，真实的密钥不会返回到前端
    copy(maskedKey);
    showSuccess(t('密钥 #{{index}} 已复制到剪贴板（脱敏版本）', { index: index + 1 }));
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'active':
        return <IconCheckCircleStroked style={{ color: '#52c41a' }} />;
      case 'disabled':
        return <IconXCircle style={{ color: '#ff4d4f' }} />;
      default:
        return <IconAlertTriangle style={{ color: '#faad14' }} />;
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'active':
        return t('正常');
      case 'disabled':
        return t('已禁用');
      default:
        return t('未知');
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'green';
      case 'disabled':
        return 'red';
      default:
        return 'orange';
    }
  };

  const getKeyTypeText = (keyType) => {
    switch (keyType) {
      case 'single':
        return t('单密钥');
      case 'multi_line':
        return t('多密钥（换行分隔）');
      case 'multi_json':
        return t('多密钥（JSON 格式）');
      default:
        return t('未知类型');
    }
  };

  const getMultiKeyModeText = (mode) => {
    switch (mode) {
      case 'random':
        return t('随机模式');
      case 'polling':
        return t('轮询模式');
      default:
        return t('单密钥模式');
    }
  };

  const handleClose = () => {
    setKeyData(null);
    onClose();
  };

  return (
    <Modal
      title={
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <IconEye style={{ marginRight: 8 }} />
          {t('查看渠道密钥')}
        </div>
      }
      visible={visible}
      onCancel={handleClose}
      width={700}
      footer={
        <Space>
          <Button onClick={handleClose}>
            {t('关闭')}
          </Button>
        </Space>
      }
      style={{ top: 20 }}
    >
      <Spin spinning={loading}>
        {keyData ? (
          <div>
            {/* 基本信息 */}
            <Card className="mb-4">
              <Descriptions
                data={[
                  { key: t('渠道名称'), value: channelName || t('未知渠道') },
                  { key: t('密钥类型'), value: getKeyTypeText(keyData.key_type) },
                  { key: t('密钥数量'), value: keyData.key_count },
                  ...(keyData.is_multi_key ? [
                    { key: t('多密钥模式'), value: getMultiKeyModeText(keyData.multi_key_mode) }
                  ] : [])
                ]}
                row
                size="small"
              />
            </Card>

            <Divider margin="12px" />

            {/* 密钥列表 */}
            <div>
              <Title heading={6} style={{ marginBottom: 12 }}>
                {t('密钥详情')}
              </Title>
              
              {keyData.keys && keyData.keys.length > 0 ? (
                <List
                  dataSource={keyData.keys}
                  renderItem={(item, index) => (
                    <List.Item
                      key={index}
                      style={{
                        padding: '12px 16px',
                        border: '1px solid var(--semi-color-border)',
                        borderRadius: '6px',
                        marginBottom: '8px',
                        backgroundColor: 'var(--semi-color-bg-1)',
                      }}
                      header={
                        <div style={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                          <div style={{ display: 'flex', alignItems: 'center', flex: 1 }}>
                            <Text strong>#{item.index + 1}</Text>
                            <div style={{ marginLeft: 12, flex: 1 }}>
                              <Text
                                code
                                style={{
                                  backgroundColor: 'var(--semi-color-fill-0)',
                                  padding: '4px 8px',
                                  borderRadius: '4px',
                                  fontSize: '12px',
                                  fontFamily: 'Monaco, Consolas, monospace',
                                }}
                              >
                                {item.masked_key}
                              </Text>
                            </div>
                          </div>
                          <Space>
                            <Tag color={getStatusColor(item.status)} size="small">
                              {getStatusIcon(item.status)}
                              <span style={{ marginLeft: 4 }}>
                                {getStatusText(item.status)}
                              </span>
                            </Tag>
                            <Tooltip content={t('复制脱敏密钥')}>
                              <Button
                                icon={<IconCopy />}
                                size="small"
                                type="tertiary"
                                onClick={() => handleCopy(item.masked_key, item.index)}
                              />
                            </Tooltip>
                          </Space>
                        </div>
                      }
                      main={
                        <div style={{ fontSize: '12px', color: 'var(--semi-color-text-2)' }}>
                          {item.last_used && (
                            <Text type="tertiary" size="small">
                              {t('最后使用: {{time}}', { time: item.last_used })}
                            </Text>
                          )}
                          {item.error_message && (
                            <div style={{ marginTop: 4 }}>
                              <Text type="danger" size="small">
                                {t('错误: {{msg}}', { msg: item.error_message })}
                              </Text>
                            </div>
                          )}
                        </div>
                      }
                    />
                  )}
                />
              ) : (
                <div style={{ textAlign: 'center', padding: 40 }}>
                  <Text type="tertiary">{t('未找到密钥信息')}</Text>
                </div>
              )}
            </div>

            {/* 安全提示 */}
            <Card
              style={{
                marginTop: 16,
                backgroundColor: 'var(--semi-color-warning-light-default)',
                border: '1px solid var(--semi-color-warning-light-active)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'flex-start' }}>
                <IconAlertTriangle
                  style={{
                    color: 'var(--semi-color-warning)',
                    marginRight: 8,
                    marginTop: 2,
                    flexShrink: 0,
                  }}
                />
                <div>
                  <Text size="small" style={{ color: 'var(--semi-color-warning-dark)' }}>
                    <strong>{t('安全提示：')}</strong>
                    {t('为了安全考虑，系统只显示脱敏后的密钥信息。完整的密钥内容不会在前端显示。')}
                  </Text>
                </div>
              </div>
            </Card>
          </div>
        ) : (
          !loading && (
            <div style={{ textAlign: 'center', padding: 40 }}>
              <Text type="tertiary">{t('无法获取密钥信息')}</Text>
            </div>
          )
        )}
      </Spin>
    </Modal>
  );
};

export default KeyViewModal;