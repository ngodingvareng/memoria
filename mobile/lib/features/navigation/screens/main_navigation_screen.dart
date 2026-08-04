import 'package:flutter/material.dart';
import 'package:hugeicons/hugeicons.dart';

import '../../home/screens/home_screen.dart';

class MainNavigationScreen extends StatefulWidget {
  const MainNavigationScreen({super.key});

  @override
  State<MainNavigationScreen> createState() => _MainNavigationScreenState();
}

class _MainNavigationScreenState extends State<MainNavigationScreen> {
  int _currentIndex = 0;

  final List<Widget> _screens = [
    const HomeScreen(),
    const Center(
      child: Text('Personal, Circle Activities'),
    ), // Placeholder Circle
    const Center(child: Text('New Capture')), // Placeholder Statistic
    const Center(child: Text('Album')), // Placeholder Statistic
    const Center(child: Text('Profile and Settings')), // Placeholder Album
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _screens[_currentIndex],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        destinations: const [
          NavigationDestination(
            icon: HugeIcon(
              icon: HugeIcons.strokeRoundedHome05,
              size: 24.0,
              strokeWidth: 2,
            ),
            selectedIcon: HugeIcon(
              icon: HugeIcons.strokeRoundedHome05,
              size: 24.0,
              strokeWidth: 2,
            ),
            label: 'Home',
          ),
          NavigationDestination(
            icon: HugeIcon(
              icon: HugeIcons.strokeRoundedActivity01,
              size: 24.0,
              strokeWidth: 2,
            ),
            selectedIcon: HugeIcon(
              icon: HugeIcons.strokeRoundedActivity01,
              size: 24.0,
              strokeWidth: 2,
            ),
            label: 'Activity',
          ),
          NavigationDestination(
            icon: HugeIcon(
              icon: HugeIcons.strokeRoundedPlusSignSquare,
              size: 24.0,
              strokeWidth: 2,
            ),
            selectedIcon: HugeIcon(
              icon: HugeIcons.strokeRoundedPlusSignSquare,
              size: 24.0,
              strokeWidth: 2,
            ),
            label: 'New',
          ),
          NavigationDestination(
            icon: HugeIcon(
              icon: HugeIcons.strokeRoundedAlbum01,
              size: 24.0,
              strokeWidth: 2,
            ),
            selectedIcon: HugeIcon(
              icon: HugeIcons.strokeRoundedAlbum01,
              size: 24.0,
              strokeWidth: 2,
            ),
            label: 'Album',
          ),
          NavigationDestination(
            icon: HugeIcon(
              icon: HugeIcons.strokeRoundedUserCircle,
              size: 24.0,
              strokeWidth: 2,
            ),
            selectedIcon: HugeIcon(
              icon: HugeIcons.strokeRoundedUserCircle,
              size: 24.0,
              strokeWidth: 2,
            ),
            label: 'Profile',
          ),
        ],
      ),
    );
  }
}
