import 'package:flutter/material.dart';
import 'package:hugeicons/hugeicons.dart';

import '../../activity/screens/capture_moment_screen.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Memoria', style: TextStyle(fontWeight: FontWeight.bold)),
        actions: [
          IconButton(
            icon: const HugeIcon(
              icon: HugeIcons.strokeRoundedMoreVertical,
              size: 24.0,
              strokeWidth: 2,
            ),
            onPressed: () {},
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16.0),
        children: [
          _buildGreetingCard(context),
          const SizedBox(height: 24),
          const Text(
            'Menunggu Konfirmasi',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),
          // Dummy data untuk aktivitas hari ini
          _buildActivityItemCard(
            context: context,
            activityName: 'Bekerja',
            time: '09:00 WIB',
            paletteColor: Colors.blue.shade100,
            isOverdue: false,
          ),
          _buildActivityItemCard(
            context: context,
            activityName: 'Olahraga Sore',
            time: '16:00 WIB',
            paletteColor: Colors.orange.shade100,
            isOverdue: true, // Contoh jika hampir lewat batas waktu
          ),
        ],
      ),
    );
  }

  Widget _buildGreetingCard(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.primaryContainer,
        borderRadius: BorderRadius.circular(16),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Halo, Kenang momenmu!', style: TextStyle(fontSize: 16)),
          SizedBox(height: 8),
          Text('Ada 2 aktivitas yang butuh konfirmasimu hari ini.',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }

  Widget _buildActivityItemCard({
    required BuildContext context,
    required String activityName,
    required String time,
    required Color paletteColor,
    required bool isOverdue,
  }) {
    return Card(
      elevation: 0,
      color: paletteColor,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        title: Text(activityName, style: const TextStyle(fontWeight: FontWeight.bold)),
        subtitle: Text(
          isOverdue ? 'Muncul sejak: $time (Segera konfirmasi)' : 'Muncul sejak: $time',
          style: TextStyle(color: isOverdue ? Colors.red.shade800 : Colors.black54),
        ),
        trailing: Checkbox(
          value: false,
          onChanged: (val) {
            // Membuka halaman "Capture Moment" sebelum benar-benar check
            Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => CaptureMomentScreen(activityName: activityName)),
            );
          },
        ),
      ),
    );
  }
}