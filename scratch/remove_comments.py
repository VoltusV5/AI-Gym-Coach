import os
import re

def remove_comments_from_text(text, ext):
    if ext in ['.go', '.js', '.vue', '.ts']:
        # Remove /* */ comments
        text = re.sub(r'/\*.*?\*/', '', text, flags=re.DOTALL)
        # Remove // comments (be careful with URLs)
        # Use a more sophisticated approach for // to avoid breaking URLs like https://
        # This regex looks for // that is NOT preceded by :
        text = re.sub(r'(?<!:)\s*//.*', '', text)
    elif ext in ['.css']:
        text = re.sub(r'/\*.*?\*/', '', text, flags=re.DOTALL)
    elif ext in ['.html']:
        text = re.sub(r'<!--.*?-->', '', text, flags=re.DOTALL)
    elif ext in ['.py', '.sql']:
        # Remove # comments for Python
        if ext == '.py':
            text = re.sub(r'#.*', '', text)
        # Remove -- comments for SQL
        if ext == '.sql':
            text = re.sub(r'--.*', '', text)
    return text

def process_file(filepath):
    ext = os.path.splitext(filepath)[1].lower()
    if ext not in ['.go', '.js', '.vue', '.css', '.html', '.sql', '.py']:
        return

    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        new_content = remove_comments_from_text(content, ext)
        
        # Remove trailing whitespace and empty lines that were just comments
        lines = [line.rstrip() for line in new_content.splitlines()]
        new_content = '\n'.join(lines)

        if content != new_content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"Cleaned: {filepath}")
    except Exception as e:
        print(f"Error processing {filepath}: {e}")

def main():
    root_dir = "."
    exclude_dirs = {'.git', 'node_modules', 'vendor', '.antigravity', '.gemini', 'dist', 'artifacts', 'scratch'}
    
    for root, dirs, files in os.walk(root_dir):
        dirs[:] = [d for d in dirs if d not in exclude_dirs]
        for file in files:
            process_file(os.path.join(root, file))

if __name__ == "__main__":
    main()
