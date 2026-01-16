import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import '../../services/api_service.dart';
import '../../theme/app_theme.dart';
import '../../widgets/stat_card.dart';
import '../../widgets/status_badge.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Dashboard'),
        actions: [
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () {},
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.refresh(dashboardStatsProvider),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.refresh(dashboardStatsProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Quick Stats
              _buildStatsGrid(context),
              const SizedBox(height: 24),
              
              // Request Rate Chart
              Text(
                'Request Rate (24h)',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                height: 200,
                child: _buildRequestChart(),
              ),
              const SizedBox(height: 24),
              
              // Recent Activity
              Text(
                'Recent Activity',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 12),
              _buildActivityList(context),
              const SizedBox(height: 24),
              
              // Service Health
              Text(
                'Service Health',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 12),
              _buildServiceHealthList(context),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStatsGrid(BuildContext context) {
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 12,
      crossAxisSpacing: 12,
      childAspectRatio: 1.5,
      children: const [
        StatCard(
          title: 'Projects',
          value: '12',
          icon: Icons.folder,
          color: AppTheme.primaryColor,
        ),
        StatCard(
          title: 'Services',
          value: '47',
          icon: Icons.cloud,
          color: AppTheme.secondaryColor,
          subtitle: '45 running',
        ),
        StatCard(
          title: 'Databases',
          value: '8',
          icon: Icons.storage,
          color: AppTheme.accentColor,
        ),
        StatCard(
          title: 'Builds Today',
          value: '23',
          icon: Icons.build,
          color: AppTheme.successColor,
          subtitle: '95% success',
        ),
      ],
    );
  }

  Widget _buildRequestChart() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: LineChart(
          LineChartData(
            gridData: FlGridData(show: false),
            titlesData: FlTitlesData(show: false),
            borderData: FlBorderData(show: false),
            lineBarsData: [
              LineChartBarData(
                spots: [
                  const FlSpot(0, 3),
                  const FlSpot(1, 4),
                  const FlSpot(2, 3.5),
                  const FlSpot(3, 5),
                  const FlSpot(4, 4),
                  const FlSpot(5, 6),
                  const FlSpot(6, 5.5),
                ],
                isCurved: true,
                color: AppTheme.primaryColor,
                barWidth: 3,
                isStrokeCapRound: true,
                dotData: FlDotData(show: false),
                belowBarData: BarAreaData(
                  show: true,
                  color: AppTheme.primaryColor.withOpacity(0.1),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildActivityList(BuildContext context) {
    return Card(
      child: Column(
        children: [
          _ActivityItem(
            icon: Icons.cloud_upload,
            title: 'api-server deployed',
            subtitle: 'production • 2 minutes ago',
            color: AppTheme.successColor,
          ),
          const Divider(height: 1),
          _ActivityItem(
            icon: Icons.build,
            title: 'frontend build completed',
            subtitle: 'staging • 5 minutes ago',
            color: AppTheme.buildingColor,
          ),
          const Divider(height: 1),
          _ActivityItem(
            icon: Icons.storage,
            title: 'Database backup completed',
            subtitle: 'production-db • 1 hour ago',
            color: AppTheme.accentColor,
          ),
          const Divider(height: 1),
          _ActivityItem(
            icon: Icons.person,
            title: 'New team member added',
            subtitle: 'john@example.com • 2 hours ago',
            color: Colors.grey,
          ),
        ],
      ),
    );
  }

  Widget _buildServiceHealthList(BuildContext context) {
    return Card(
      child: Column(
        children: [
          _ServiceHealthItem(
            name: 'api-server',
            status: 'running',
            cpu: 45,
            memory: 60,
          ),
          const Divider(height: 1),
          _ServiceHealthItem(
            name: 'frontend',
            status: 'running',
            cpu: 20,
            memory: 35,
          ),
          const Divider(height: 1),
          _ServiceHealthItem(
            name: 'worker',
            status: 'running',
            cpu: 80,
            memory: 70,
          ),
          const Divider(height: 1),
          _ServiceHealthItem(
            name: 'cron-jobs',
            status: 'stopped',
            cpu: 0,
            memory: 0,
          ),
        ],
      ),
    );
  }
}

class _ActivityItem extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final Color color;

  const _ActivityItem({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: color.withOpacity(0.1),
        child: Icon(icon, color: color, size: 20),
      ),
      title: Text(title),
      subtitle: Text(subtitle),
      dense: true,
    );
  }
}

class _ServiceHealthItem extends StatelessWidget {
  final String name;
  final String status;
  final int cpu;
  final int memory;

  const _ServiceHealthItem({
    required this.name,
    required this.status,
    required this.cpu,
    required this.memory,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Expanded(
            flex: 2,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(name, style: const TextStyle(fontWeight: FontWeight.w500)),
                const SizedBox(height: 4),
                StatusBadge(status: status),
              ],
            ),
          ),
          Expanded(
            child: Column(
              children: [
                Text('CPU', style: Theme.of(context).textTheme.bodySmall),
                Text('$cpu%', style: const TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
          ),
          Expanded(
            child: Column(
              children: [
                Text('Memory', style: Theme.of(context).textTheme.bodySmall),
                Text('$memory%', style: const TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// Provider placeholder
final dashboardStatsProvider = FutureProvider((ref) async {
  return {};
});
