import 'package:flutter/material.dart';

class CaptureMomentScreen extends StatefulWidget {
  final String activityName;
  const CaptureMomentScreen({super.key, required this.activityName});

  @override
  State<CaptureMomentScreen> createState() => _CaptureMomentScreenState();
}

class _CaptureMomentScreenState extends State<CaptureMomentScreen> {
  bool _isCompleted = true; // Was the activity done, or skipped?

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Confirm: ${widget.activityName}'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Toggle Done / Not done
            SegmentedButton<bool>(
              segments: const [
                ButtonSegment(value: true, label: Text('Done'), icon: Icon(Icons.check_circle)),
                ButtonSegment(value: false, label: Text('Skipped'), icon: Icon(Icons.cancel)),
              ],
              selected: {_isCompleted},
              onSelectionChanged: (Set<bool> newSelection) {
                setState(() {
                  _isCompleted = newSelection.first;
                });
              },
            ),
            const SizedBox(height: 24),

            // Journal / Textarea
            const Text('Log the moment (Optional)', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            TextField(
              maxLines: 5,
              decoration: InputDecoration(
                hintText: _isCompleted
                    ? 'What stood out about this activity?'
                    : 'Why was this activity missed?',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                filled: true,
                fillColor: Colors.grey.shade50,
              ),
            ),
            const SizedBox(height: 24),

            // Photo upload
            const Text('Photo Documentation', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            InkWell(
              onTap: () {
                // Multiple image picker logic
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
                    Text('Add Photo (Multiple allowed)', style: TextStyle(color: Colors.grey)),
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
            // Save data to the database and return to Home
            Navigator.pop(context);
          },
          style: FilledButton.styleFrom(
            padding: const EdgeInsets.symmetric(vertical: 16),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
          child: const Text('Save Moment', style: TextStyle(fontSize: 16)),
        ),
      ),
    );
  }
}