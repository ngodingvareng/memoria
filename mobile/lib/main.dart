import 'package:flutter/material.dart';
import 'features/navigation/screens/main_navigation_screen.dart';

void main() {
  runApp(const NostalgiaApp());
}

class NostalgiaApp extends StatelessWidget {
  const NostalgiaApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Nostalgia Tracker',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.teal, // A color that evokes calm/journaling
          brightness: Brightness.light,
        ),
        useMaterial3: true,
      ),
      home: const MainNavigationScreen(),
    );
  }
}