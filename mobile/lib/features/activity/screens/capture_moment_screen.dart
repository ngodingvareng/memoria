import 'package:flutter/material.dart';

class CaptureMomentScreen extends StatefulWidget {
  final String activityName;
  const CaptureMomentScreen({super.key, required this.activityName});

  @override
  State<CaptureMomentScreen> createState() => _CaptureMomentScreenState();
}

class _CaptureMomentScreenState extends State<CaptureMomentScreen> {
  bool _isCompleted = true; // Apakah aktivitas dilakukan atau di-skip?

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Konfirmasi: ${widget.activityName}'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Toggle Dilakukan / Tidak Dilakukan
            SegmentedButton<bool>(
              segments: const [
                ButtonSegment(value: true, label: Text('Dilaksanakan'), icon: Icon(Icons.check_circle)),
                ButtonSegment(value: false, label: Text('Dilewatkan'), icon: Icon(Icons.cancel)),
              ],
              selected: {_isCompleted},
              onSelectionChanged: (Set<bool> newSelection) {
                setState(() {
                  _isCompleted = newSelection.first;
                });
              },
            ),
            const SizedBox(height: 24),
            
            // Jurnal / Textarea
            const Text('Catat momennya (Opsional)', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            TextField(
              maxLines: 5,
              decoration: InputDecoration(
                hintText: _isCompleted 
                    ? 'Apa yang berkesan dari aktivitas ini?' 
                    : 'Kenapa aktivitas ini terlewatkan?',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                filled: true,
                fillColor: Colors.grey.shade50,
              ),
            ),
            const SizedBox(height: 24),

            // Upload Gambar
            const Text('Dokumentasi Foto', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            InkWell(
              onTap: () {
                // Logika pick image multiple
              },
              child: Container(
                height: 100,
                width: double.infinity,
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey.shade400, style: BorderStyle.solid),
                  borderRadius: BorderRadius.circular(12),
                  color: Colors.grey.shade100,
                ),
                child: const Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.add_a_photo, color: Colors.grey),
                    SizedBox(height: 8),
                    Text('Tambah Foto (Bisa lebih dari 1)', style: TextStyle(color: Colors.grey)),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: Padding(
        padding: const EdgeInsets.all(16.0),
        child: FilledButton(
          onPressed: () {
            // Save data ke database dan kembali ke Home
            Navigator.pop(context);
          },
          style: FilledButton.styleFrom(
            padding: const EdgeInsets.symmetric(vertical: 16),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
          child: const Text('Simpan Momen', style: TextStyle(fontSize: 16)),
        ),
      ),
    );
  }
}