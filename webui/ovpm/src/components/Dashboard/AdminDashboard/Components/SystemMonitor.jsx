import React from 'react';

class SystemMonitor extends React.Component {
  formatBytes = (bytes) => {
    if (bytes >= 1024 * 1024 * 1024) {
      return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    } else if (bytes >= 1024 * 1024) {
      return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
    } else if (bytes >= 1024) {
      return (bytes / 1024).toFixed(2) + ' KB';
    }
    return bytes + ' B';
  };

  formatPercentage = (value) => {
    return (value * 100).toFixed(1) + '%';
  };

  renderProgressBar = (percentage, color = '#4CAF50') => {
    return (
      <div style={{
        width: '100%',
        backgroundColor: '#e0e0e0',
        borderRadius: '5px',
        overflow: 'hidden'
      }}>
        <div style={{
          width: `${percentage * 100}%`,
          backgroundColor: color,
          height: '20px',
          transition: 'width 0.3s ease'
        }}></div>
      </div>
    );
  };

  render() {
    const { systemStatus } = this.props;

    if (!systemStatus) {
      return <div>Loading system status...</div>;
    }

    return (
      <div style={{ padding: '10px', background: '#f5f5f5', borderRadius: '5px' }}>
        <h4 style={{ margin: '0 0 10px 0' }}>System Status</h4>

        <div style={{ marginBottom: '10px' }}>
          <div><strong>Load Average:</strong></div>
          <div>1min: {systemStatus.load_average[0]}</div>
          <div>5min: {systemStatus.load_average[1]}</div>
          <div>15min: {systemStatus.load_average[2]}</div>
        </div>

        <div style={{ marginBottom: '10px' }}>
          <div><strong>CPU Usage:</strong> {this.formatPercentage(systemStatus.cpu_usage)}</div>
          {this.renderProgressBar(systemStatus.cpu_usage, '#2196F3')}
        </div>

        <div style={{ marginBottom: '10px' }}>
          <div><strong>Memory:</strong> {this.formatBytes(systemStatus.memory_used)} / {this.formatBytes(systemStatus.memory_total)}</div>
          {this.renderProgressBar(systemStatus.memory_used / systemStatus.memory_total, '#4CAF50')}
        </div>

        <div style={{ marginBottom: '10px' }}>
          <div><strong>Swap:</strong> {this.formatBytes(systemStatus.swap_used)} / {this.formatBytes(systemStatus.swap_total)}</div>
          {this.renderProgressBar(systemStatus.swap_used / systemStatus.swap_total, '#FF9800')}
        </div>

        {systemStatus.disk_usage && systemStatus.disk_usage.map((disk, index) => (
          <div key={index} style={{ marginBottom: '5px' }}>
            <div><strong>Disk {disk.mount}:</strong> {this.formatPercentage(disk.used_percentage)}</div>
            {this.renderProgressBar(disk.used_percentage, '#9C27B0')}
          </div>
        ))}
      </div>
    );
  }
}

export default SystemMonitor;