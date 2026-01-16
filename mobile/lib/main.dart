import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'screens/splash_screen.dart';
import 'screens/auth/login_screen.dart';
import 'screens/auth/register_screen.dart';
import 'screens/home/dashboard_screen.dart';
import 'screens/projects/projects_screen.dart';
import 'screens/projects/project_detail_screen.dart';
import 'screens/services/services_screen.dart';
import 'screens/services/service_detail_screen.dart';
import 'screens/services/service_logs_screen.dart';
import 'screens/databases/databases_screen.dart';
import 'screens/databases/database_detail_screen.dart';
import 'screens/clusters/clusters_screen.dart';
import 'screens/clusters/cluster_detail_screen.dart';
import 'screens/builds/builds_screen.dart';
import 'screens/builds/build_detail_screen.dart';
import 'screens/deployments/deployments_screen.dart';
import 'screens/monitoring/monitoring_screen.dart';
import 'screens/settings/settings_screen.dart';
import 'screens/settings/profile_screen.dart';
import 'screens/settings/notifications_screen.dart';
import 'theme/app_theme.dart';

void main() {
  runApp(const ProviderScope(child: NorthStackApp()));
}

class NorthStackApp extends ConsumerWidget {
  const NorthStackApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp.router(
      title: 'NorthStack',
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: ThemeMode.system,
      routerConfig: _router,
      debugShowCheckedModeBanner: false,
    );
  }
}

final _router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      builder: (context, state) => const SplashScreen(),
    ),
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: '/register',
      builder: (context, state) => const RegisterScreen(),
    ),
    ShellRoute(
      builder: (context, state, child) => MainShell(child: child),
      routes: [
        GoRoute(
          path: '/dashboard',
          builder: (context, state) => const DashboardScreen(),
        ),
        GoRoute(
          path: '/projects',
          builder: (context, state) => const ProjectsScreen(),
          routes: [
            GoRoute(
              path: ':id',
              builder: (context, state) => ProjectDetailScreen(
                projectId: state.pathParameters['id']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: '/services',
          builder: (context, state) => const ServicesScreen(),
          routes: [
            GoRoute(
              path: ':id',
              builder: (context, state) => ServiceDetailScreen(
                serviceId: state.pathParameters['id']!,
              ),
              routes: [
                GoRoute(
                  path: 'logs',
                  builder: (context, state) => ServiceLogsScreen(
                    serviceId: state.pathParameters['id']!,
                  ),
                ),
              ],
            ),
          ],
        ),
        GoRoute(
          path: '/databases',
          builder: (context, state) => const DatabasesScreen(),
          routes: [
            GoRoute(
              path: ':id',
              builder: (context, state) => DatabaseDetailScreen(
                databaseId: state.pathParameters['id']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: '/clusters',
          builder: (context, state) => const ClustersScreen(),
          routes: [
            GoRoute(
              path: ':id',
              builder: (context, state) => ClusterDetailScreen(
                clusterId: state.pathParameters['id']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: '/builds',
          builder: (context, state) => const BuildsScreen(),
          routes: [
            GoRoute(
              path: ':id',
              builder: (context, state) => BuildDetailScreen(
                buildId: state.pathParameters['id']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: '/deployments',
          builder: (context, state) => const DeploymentsScreen(),
        ),
        GoRoute(
          path: '/monitoring',
          builder: (context, state) => const MonitoringScreen(),
        ),
        GoRoute(
          path: '/settings',
          builder: (context, state) => const SettingsScreen(),
          routes: [
            GoRoute(
              path: 'profile',
              builder: (context, state) => const ProfileScreen(),
            ),
            GoRoute(
              path: 'notifications',
              builder: (context, state) => const NotificationsScreen(),
            ),
          ],
        ),
      ],
    ),
  ],
);

class MainShell extends StatelessWidget {
  final Widget child;

  const MainShell({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.dashboard_outlined),
            selectedIcon: Icon(Icons.dashboard),
            label: 'Dashboard',
          ),
          NavigationDestination(
            icon: Icon(Icons.folder_outlined),
            selectedIcon: Icon(Icons.folder),
            label: 'Projects',
          ),
          NavigationDestination(
            icon: Icon(Icons.cloud_outlined),
            selectedIcon: Icon(Icons.cloud),
            label: 'Services',
          ),
          NavigationDestination(
            icon: Icon(Icons.storage_outlined),
            selectedIcon: Icon(Icons.storage),
            label: 'Databases',
          ),
          NavigationDestination(
            icon: Icon(Icons.settings_outlined),
            selectedIcon: Icon(Icons.settings),
            label: 'Settings',
          ),
        ],
        selectedIndex: _getSelectedIndex(context),
        onDestinationSelected: (index) => _onDestinationSelected(context, index),
      ),
    );
  }

  int _getSelectedIndex(BuildContext context) {
    final location = GoRouterState.of(context).uri.path;
    if (location.startsWith('/dashboard')) return 0;
    if (location.startsWith('/projects')) return 1;
    if (location.startsWith('/services')) return 2;
    if (location.startsWith('/databases')) return 3;
    if (location.startsWith('/settings')) return 4;
    return 0;
  }

  void _onDestinationSelected(BuildContext context, int index) {
    switch (index) {
      case 0:
        context.go('/dashboard');
        break;
      case 1:
        context.go('/projects');
        break;
      case 2:
        context.go('/services');
        break;
      case 3:
        context.go('/databases');
        break;
      case 4:
        context.go('/settings');
        break;
    }
  }
}
