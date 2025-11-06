import React from 'react';
import { Line } from 'react-chartjs-2';
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler
} from 'chart.js';

ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler
);

class NetworkMonitor extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            selectedInterface: '',
            currentStats: null,
            trafficHistory: {
                labels: [],
                upload: [],
                download: []
            }
        };
        this.historySize = 60; // 60 точек = 1 минута
    }

    componentDidMount() {
        this.props.onInterfaceChange(this.state.selectedInterface);
    }

    handleInterfaceChange = (e) => {
        const interfaceName = e.target.value;
        this.setState({ selectedInterface: interfaceName }, () => {
            this.props.onInterfaceChange(interfaceName);
            this.resetHistory();
        });
    };

    resetHistory = () => {
        this.setState({
            trafficHistory: {
                labels: Array(this.historySize).fill(''),
                upload: Array(this.historySize).fill(0),
                download: Array(this.historySize).fill(0)
            }
        });
    };

    updateStats = (stats) => {
        if (!stats) {
            console.warn("updateStats called with undefined stats");
            return;
        }

        // Конвертируем ВСЕ строковые числа в настоящие числа
        const safeStats = {
            interface_name: stats.interface_name,
            tx_bytes: Number(stats.tx_bytes) || 0,
            rx_bytes: Number(stats.rx_bytes) || 0,
            tx_bytes_per_sec: Number(stats.tx_bytes_per_sec) || 0,
            rx_bytes_per_sec: Number(stats.rx_bytes_per_sec) || 0,
            tx_packets: Number(stats.tx_packets) || 0,
            rx_packets: Number(stats.rx_packets) || 0,
            timestamp: stats.timestamp
        };

        console.log("Processed stats:", safeStats); // Для отладки

        this.setState({ currentStats: safeStats });

        // Обновление истории (убедитесь что используете числа)
        this.setState(prevState => {
            const newLabels = [...prevState.trafficHistory.labels];
            const newUpload = [...prevState.trafficHistory.upload];
            const newDownload = [...prevState.trafficHistory.download];

            newLabels.push(new Date().toLocaleTimeString());
            newUpload.push(safeStats.tx_bytes_per_sec / 1024 / 1024); // MB/s
            newDownload.push(safeStats.rx_bytes_per_sec / 1024 / 1024); // MB/s

            if (newLabels.length > this.historySize) {
                newLabels.shift();
                newUpload.shift();
                newDownload.shift();
            }

            return {
                trafficHistory: {
                    labels: newLabels,
                    upload: newUpload,
                    download: newDownload
                }
            };
        });
    };

    getChartData = () => {
        return {
            labels: this.state.trafficHistory.labels,
            datasets: [
                {
                    label: 'Upload (MB/s)',
                    data: this.state.trafficHistory.upload,
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    tension: 0.4,
                    fill: true
                },
                {
                    label: 'Download (MB/s)',
                    data: this.state.trafficHistory.download,
                    borderColor: 'rgb(54, 162, 235)',
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    tension: 0.4,
                    fill: true
                }
            ]
        };
    };

    getChartOptions = () => {
        return {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Network Traffic (Last 60 seconds)'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'MB/s'
                    }
                }
            }
        };
    };

    formatSpeed = (bytesPerSec) => {
        if (bytesPerSec >= 1024 * 1024 * 1024) {
            return (bytesPerSec / (1024 * 1024 * 1024)).toFixed(2) + ' GB/s';
        } else if (bytesPerSec >= 1024 * 1024) {
            return (bytesPerSec / (1024 * 1024)).toFixed(2) + ' MB/s';
        } else if (bytesPerSec >= 1024) {
            return (bytesPerSec / 1024).toFixed(2) + ' KB/s';
        }
        return bytesPerSec.toFixed(2) + ' B/s';
    };

    render() {
        const { interfaces } = this.props;
        const { selectedInterface, currentStats } = this.state;

        return (
            <div style={{ padding: '10px', background: '#f5f5f5', borderRadius: '5px' }}>
                <div style={{ marginBottom: '10px' }}>
                    <select
                        value={selectedInterface}
                        onChange={this.handleInterfaceChange}
                        style={{ padding: '5px', width: '200px' }}
                    >
                        <option value="">Select Interface</option>
                        {interfaces.map(iface => (
                            <option key={iface.name} value={iface.name}>
                                {iface.name} ({iface.ip || 'No IP'})
                            </option>
                        ))}
                    </select>
                </div>

                {currentStats && (
                    <div style={{ marginBottom: '10px' }}>
                        <div><strong>Upload:</strong> {this.formatSpeed(currentStats.tx_bytes_per_sec)}</div>
                        <div><strong>Download:</strong> {this.formatSpeed(currentStats.rx_bytes_per_sec)}</div>
                        <div><strong>Total:</strong> {this.formatSpeed(currentStats.tx_bytes_per_sec + currentStats.rx_bytes_per_sec)}</div>
                    </div>
                )}

                {selectedInterface && (
                    <div style={{ height: '200px' }}>
                        <Line data={this.getChartData()} options={this.getChartOptions()} />
                    </div>
                )}
            </div>
        );
    }
}

export default NetworkMonitor;
